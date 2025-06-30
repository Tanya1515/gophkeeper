package client

import (
	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
	"go.uber.org/zap"
)

// User login
var User string

// Client - type, that constaints necessary data for client usage.
type Client struct {
	ClientStorage cs.ClientStorage  // Local client storage
	ClientLogger  zap.SugaredLogger // Logger saves all server info
}

// UserData - type, that containts all necessary information about
// user and sensetive data identificator.
type UserData struct {
	user              string // User login
	JWTtoken          string // User JWT token for authentification
	dataIdentificator string // Identificator for sensetive data: file name, bank card number or application name, for which password is used.
}

// OperationResult - type, that containts data for detecting result after
// executing operation repeatedly.
type OperationResult struct {
	operationID string // Unigue identificator for operation
	result      string // Result of the executed operation
}
