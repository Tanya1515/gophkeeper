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

	GetBankCard(cardNumber, userName string) (cvc, date, bankName, metadatabankCard string, err error)
	GetFile(fileName, userName string) (string, string, error)
	GetPassword(application, userName string) (password, metadata string, err error)

	DeleteBankCard(cardNumber, userName string) error
	DeleteFile(fileName, userName string) error
	DeletePassword(application, userName string) error

	UploadBankCard(cardNumber, cvc, date, bankName, metadatabankCard, uploadTime, userName string) error
	UploadFile(fileName, filePath, metadata, uploadTime, userName string) error
	UploadPassword(application, password, metadata, uploadTime, userName string) error

	SaveCardOperation(cardNumber, userName string, operation Operation, fields []string, time string) error
	SaveFileOperation(fileName, userName string, operation Operation, fields []string, time string) error
	SavePasswordOperation(application, userName string, operation Operation, fields []string, time string) error

	GetAllCardOperation() (map[string][]string, error)
	GetAllFileOperation() (map[string][]string, error)
	GetAllPasswordOperation() (map[string][]string, error)
}
