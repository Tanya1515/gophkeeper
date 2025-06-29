// Client_storage - package, that contains interface for
// client storage. Client storage can act as cache system
// for users' sensetive data or system for saving user requests
// if server is not available.
package client_storage

// ClientStorage - interface for describeng storage for caching
// sensetive data on client side.
type ClientStorage interface {
	Connect() error

	CreateUser(userName string) error

	GetBankCard(cardNumber, userName string) (cvc, date, bankName, metadatabankCard string, exists bool, err error)
	GetFile(fileName, userName string) (string, string, bool, []byte, error)
	GetPassword(application, userName string) (password, metadata, uploadTime string, exists bool, err error)

	DeleteBankCard(cardNumber, userName string) error
	DeleteFile(fileName, userName string) error
	DeletePassword(application, userName string) error

	UploadBankCard(cardNumber, cvc, date, bankName, metadatabankCard, uploadTime, userName string) error
	UploadFile(fileName, filePath, metadata, uploadTime, userName string) error
	UploadPassword(application, password, metadata, uploadTime, userName string) error

	SaveCardOperation(cardNumber, userName string, operation Operation, fields []string, time string) error
	SaveFileOperation(fileName, userName string, operation Operation, fields []string, time, filePath string) error
	SavePasswordOperation(application, userName string, operation Operation, fields []string, time string) error

	GetAllCardWithOperation() (map[string][]string, error)
	GetAllFileWithOperation() (map[string][]string, error)
	GetAllPasswordWithOperation() (map[string][]string, error)

	GetPasswordOperationsInfo(user, application string) (map[string]OperationInfo, error)
	GetCardOperationsInfo(user, cardNumber string) (map[string]OperationInfo, error)
	GetFileOpearionsInfo(user, fileName string) (map[string]OperationInfo, error)

	DeleteOperaionByID(operationID, dataType string) error
}
