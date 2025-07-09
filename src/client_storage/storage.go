// Client_storage - package, that contains interface for
// client storage. Client storage can act as cache system
// for users' sensetive data or system for saving user requests
// if server is not available.
package client_storage

import ut "github.com/Tanya1515/gophkeeper.git/src/utils"

// ClientStorage - interface for describeng storage for caching
// sensetive data on client side.
type ClientStorage interface {
	Connect() error

	CreateUser(userName string) error
	UpdateCacheMiss(userName string, cacheMiss int) error
	GetCacheMiss(userName string) (int, error)

	GetBankCard(cardNumber, userName string) (cvc, date, bankName, metadatabankCard string, initVector []byte, exists bool, err error)
	GetFile(fileName, userName string) (string, string, bool, []byte, error)
	GetPassword(application, userName string) (password, metadata, uploadTime string, initVector []byte, exists bool, err error)

	DeleteBankCard(cardNumber, userName string) error
	DeleteFile(fileName, userName string) (string, error)
	DeletePassword(application, userName string) error

	UploadBankCard(cardNumber, cvc, date, bankName, metadatabankCard, uploadTime, lastUpdated, userName string, initVector []byte) error
	UploadFile(fileName, filePath, metadata, uploadTime, lastUpdated, userName string) error
	UploadPassword(application, password, metadata, uploadTime, lastUpdated, userName string, initVector []byte) error

	SaveCardOperation(cardNumber, userName string, operation ut.Operation, fields []string, time string) error
	SaveFileOperation(fileName, userName string, operation ut.Operation, fields []string, time, filePath string) error
	SavePasswordOperation(application, userName string, operation ut.Operation, fields []string, time string) error

	GetAllCardWithOperation() (map[string][]string, error)
	GetAllFileWithOperation() (map[string][]string, error)
	GetAllPasswordWithOperation() (map[string][]string, error)

	GetPasswordOperationsInfo(user, application string) ([]ut.OperationInfo, error)
	GetCardOperationsInfo(user, cardNumber string) ([]ut.OperationInfo, error)
	GetFileOpearionsInfo(user, fileName string) ([]ut.OperationInfo, error)

	ClearPasswordsByDate(userName string) error
	ClearFilesByDate(userName string) ([]string, error)
	ClearCardsByDate(userName string) error

	DeleteOperaionByID(operationID, dataType string) error
}
