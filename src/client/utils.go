package client

import (
	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
)

var User string

type Client struct {
	ClientStorage cs.ClientStorage
}

type UserData struct {
	user              string
	JWTtoken          string
	dataIdentificator string
}
