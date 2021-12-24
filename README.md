---start oauth server

go run server.go


---Step1: authorize request
GET
http://localhost:9096/oauth/authorize?response_type=code&client_id=222222&state=xyz&redirect_uri=http%3A%2F%2Flocalhost%3A9094%2Foauth2&scope=all


---Step2: 
Use the redirect uri to get the "code" parameter( {{code}} ) in the URL


---Step3: token request
POST
http://localhost:9096/oauth/token

header Basic Auth:
    Username=222222
    Password=22222222

x-www-form-urlencoded parameters:
    code={{code}}
    grant_type=authorization_code
    redirect_uri=http%3A%2F%2Flocalhost%3A9094%2Foauth2
