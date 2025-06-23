// Client_storage - package, that contains interface for
// client storage. Client storage can act as cache system
// for users' sensetive data or system for saving user requests
// if server is not available.
package client_storage

// Custom type, that describes operations in application
type Operation string

// Constants, that are related to type Operations.
// The values describe CRUD operations with objects.
const (
	Get    Operation = "get"
	Update Operation = "update"
	Delete Operation = "delete"
	Create Operation = "create"
)

// ClientStorage - interface for describeng storage for caching
// sensetive data on client side.
type ClientStorage interface {
	Connect() error

	GetBankCard(cardNumber string, userID int) (cvc, date, bankName, metadatabankCard string, err error)
	GetFile(fileName string, userID int) (string, string, error)
	GetPassword(application string, userID int) (password, metadata string, err error)

	DeleteBankCard(cardNumber string, userID int) error
	DeleteFile(fileName string, userID int) error
	DeletePassword(application string, userID int) error

	UploadBankCard(cardNumber, cvc, date, bankName, metadatabankCard, uploadTime string, userID int) error
	UploadFile(fileName, filePath, metadata, uploadTime string, userID int) error
	UploadPassword(application, password, metadata, uploadTime string, userID int) error

	SaveCardOperation(cardNumber string, operation Operation, fields []string, time string, userID int) error
	SaveFileOperation(fileName string, operation Operation, fields []string, time string, userID int) error
	SavePasswordOperation(application string, operation Operation, fields []string, time string, userID int) error
}
