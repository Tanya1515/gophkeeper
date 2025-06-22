//Client_storage - package, that contains interface for 
// client storage. Client storage can act as cache system
// for users' sensetive data or system for saving user requests
// if server is not available.
package client_storage 


type Operation string

const (
	Get Operation = "get"
	Update Operation = "update"
	Delete Operation = "delete"
	Create Operation = "create"
)

// ClientStorage - interface for describeng storage for caching 
// sensetive data on client side.
type ClientStorage interface {
	Connect() 

	GetBankCard(cardNumber string) (password string, metadata string, err error)
	GetFile(fileName string) error
	GetPassword(application string) error

	DeleteBankCard(cardNumber string) error
	DeleteFile(fileName string) error
	DeletePassword(application string) error

	UploadBankCard(cardNumber string) error
	UploadFile(fileName, filePath, metadata string) error
	UploadPassword(application, password, metadata string) error

	SaveCardOperation(cardNumber string, operation Operation, fields []string, time string) error
	SaveFileOperation(fileName string, operation Operation, fields []string, time string) error
	SavePasswordOperation(application string, operation Operation, fields []string, time string) error

}