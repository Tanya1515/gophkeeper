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

	GetBankCard(cardNumber string, userID int) (password string, metadata string, err error)
	GetFile(fileName string, userID int) error
	GetPassword(application string, userID int) error

	DeleteBankCard(cardNumber string, userID int) error
	DeleteFile(fileName string, userID int) error
	DeletePassword(application string, userID int) error

	UploadBankCard(cardNumber, cvc, date, bankName, metadatabankCard, uploadTime string, userID int) error
	UploadFile(fileName, filePath, metadata string, userID int) error
	UploadPassword(application, password, metadata string, userID int) error

	SaveCardOperation(cardNumber string, operation Operation, fields []string, time string) error
	SaveFileOperation(fileName string, operation Operation, fields []string, time string) error
	SavePasswordOperation(application string, operation Operation, fields []string, time string) error

}