---Start oAuth server

    go run server.go


---Step1: client sends the following authorization request to authorization server to activate the oAuth 2.0 process 

---Note: log in by using username: timo123 password: 123123 

    GET

    http://localhost:9096/oauth/authorize?response_type=code&client_id=222222&state=xyz&redirect_uri=http%3A%2F%2Flocalhost%3A9094%2Foauth2&scope=openid&nonce=cdef


---Step2: 

    clinet utilizes the redirect uri to get the "code" parameter( {{code}} ) in the URL


---Step3: client sends following token request to authorization owner in order to obtain the access token for resource owner

    POST

    http://localhost:9096/oauth/token

    header Basic Auth:
        Username=222222
        Password=22222222

    x-www-form-urlencoded parameters:

        code={{code}}

        grant_type=authorization_code
    
        redirect_uri=http://localhost:9094/oauth2

---Step4: client sends following token request to authorization owner in order to refresh the access token for resource owner

    POST

    http://localhost:9096/oauth/token

    header Basic Auth:
        Username=222222
        Password=22222222

    x-www-form-urlencoded parameters:

        grant_type=refresh_token
    
        refresh_token={{refresh_token}}
