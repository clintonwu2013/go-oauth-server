---start oauth server
go run server.go


---authorize
http://localhost:9096/oauth/authorize?response_type=code&client_id=222222&state=xyz&redirect_uri=http%3A%2F%2Flocalhost%3A9094%2Foauth2&scope=all


---token

client_id=222222 client_secret=22222222 ----->帶在header basic authorization內

http://localhost:9096/oauth/token?code=NZI1ZJBJNDUTMMM2YS0ZNTNKLTG0NTYTYJUZN2JMYJI3ZJRL&grant_type=authorization_code&redirect_uri=http%3A%2F%2Flocalhost%3A9094%2Foauth2&scope=all
