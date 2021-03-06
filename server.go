package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	"github.com/go-session/session"
	"github.com/golang-jwt/jwt"
)

var (
	portvar               = 9096
	client_id             = "222222"
	client_secret         = "22222222"
	doamin                = "http://localhost:9094"
	redirect_uri          = "http://localhost:9094/oauth2"
	code_challenge        = "Qn3Kywp0OiU4NK_AFzGPlmrcYJDJ13Abj_jdL08Ahg8"
	scope                 = "all"
	code_challenge_method = "S256"
	state                 = "xyz"
	jwtSecret             = []byte("jwtSecret_changingtec")
)

type Claims struct {
	UID    string `json:"UID"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Email  string `json:"email"`
	Phone  string `json:"phone"`
	Group  string `json:"group"`
	Nonce  string `json:"nonce,omitempty"`
	jwt.StandardClaims
}

func main() {

	manager := manage.NewDefaultManager()
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	// token store
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	manager.MapAccessGenerate(generates.NewAccessGenerate())

	clientStore := store.NewClientStore()
	clientStore.Set(client_id, &models.Client{
		ID:     client_id,
		Secret: client_secret,
	})
	manager.MapClientStorage(clientStore)

	srv := server.NewServer(server.NewConfig(), manager)

	srv.SetUserAuthorizationHandler(userAuthorizeHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/auth", authHandler)

	http.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		store, err := session.Start(r.Context(), w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var form url.Values
		if v, ok := store.Get("ClientData"); ok {
			form = v.(url.Values)
		}
		r.Form = form
		store.Delete("ClientData")
		store.Save()

		err = srv.HandleAuthorizeRequest(w, r)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {

		grantType := r.FormValue("grant_type")
		if grantType != "authorization_code" && grantType != "refresh_token" {
			http.Error(w, "accept only grant type: authorization_code, refresh_token", http.StatusInternalServerError)
			return
		}
		ctx := r.Context()

		gt, tgr, err := srv.ValidationTokenRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ti, err := srv.GetAccessToken(ctx, gt, tgr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tokenData := srv.GetTokenData(ti)

		//produce id token
		uidTmp := ti.GetUserID()
		uidAndNonce := strings.Split(uidTmp, "_")
		fmt.Println("##### uidAndNonce=", uidAndNonce)
		var uid string
		var nonce string
		for i := range uidAndNonce {
			if i == 0 {
				uid = uidAndNonce[i]
			} else {
				nonce = uidAndNonce[i]
			}
		}
		now := time.Now()
		claims := Claims{
			UID:    uid,
			Name:   "TestUser",
			Domain: "ad2008.vm",
			Email:  "timo123@changingtec.com",
			Group:  "AllUser,ad2008AD",
			Phone:  "0912345678",
			Nonce:  nonce,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: now.Add(7200 * time.Second).Unix(),
				IssuedAt:  now.Unix(),
				Issuer:    "changingtec.com",
			},
		}
		idTokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		idToken, err := idTokenClaims.SignedString(jwtSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tokenData["id_token"] = idToken

		fmt.Println("tokenData[access_token]=", tokenData["access_token"])
		fmt.Println("tokenData[expires_in]=", tokenData["expires_in"])

		err = returnAccessTokenData(w, tokenData, nil)
		//err := srv.HandleTokenRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {

		token, err := srv.ValidationBearerToken(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		data := map[string]interface{}{
			"expires_in": int64(token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Sub(time.Now()).Seconds()),
			"client_id":  token.GetClientID(),
			"user_id":    token.GetUserID(),
		}
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		e.Encode(data)
	})

	log.Printf("Server is running at %d port.\n", portvar)
	log.Printf("Point your OAuth client Auth endpoint to %s:%d%s", "http://localhost", portvar, "/oauth/authorize")
	log.Printf("Point your OAuth client Token endpoint to %s:%d%s", "http://localhost", portvar, "/oauth/token")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", portvar), nil))
}

func returnAccessTokenData(w http.ResponseWriter, data map[string]interface{}, header http.Header, statusCode ...int) error {

	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	for key := range header {
		w.Header().Set(key, header.Get(key))
	}

	status := http.StatusOK
	if len(statusCode) > 0 && statusCode[0] > 0 {
		status = statusCode[0]
	}

	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		return
	}
	scope := r.FormValue("scope")
	fmt.Println("####### scope=", scope)
	if !strings.Contains(scope, "openid") {
		http.Error(w, "accept only scope=openid", http.StatusBadRequest)
		return
	}
	uid, ok := store.Get("LoggedInUserID")
	if !ok {
		if r.Form == nil {
			r.ParseForm()
		}
		store.Set("ClientData", r.Form)
		store.Save()
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
	nonce := r.FormValue("nonce")
	fmt.Println("####### nonce=", nonce)
	userID = uid.(string) + "_" + nonce
	store.Delete("LoggedInUserID")
	store.Save()
	return
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Method == "POST" {
		//TODO: VERIFY IN ide AND GET UID
		var req struct {
			Username string
			Password string
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Println("##### verify requset=", req)
		username := req.Username
		password := req.Password
		var data struct {
			ErrorCode int    `json:"errorCode"`
			ErrorMsg  string `json:"errorMsg"`
		}
		if !verifyUser(username, password) {

			data.ErrorCode = -1
			data.ErrorMsg = "invalid username or password !!!"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(data)
			return
		}

		uid := "1"
		store.Set("LoggedInUserID", uid)
		store.Save()
		data.ErrorCode = 0
		data.ErrorMsg = "verify passed!!!"
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(data)

		return
	}
	outputHTML(w, r, "static/login.html")
}

func verifyUser(username, password string) (result bool) {
	if username == "timo123" && password == "123123" {
		result = true
		return
	}

	result = false
	return
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, ok := store.Get("LoggedInUserID"); !ok {
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
		return
	}
	outputHTML(w, r, "static/auth.html")
}

func outputHTML(w http.ResponseWriter, req *http.Request, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer file.Close()
	fi, _ := file.Stat()
	http.ServeContent(w, req, file.Name(), fi.ModTime(), file)
}
