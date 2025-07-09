// Datastorage - package, that describes interface for using
// database in the context of sensetive data storage.
package datastorage

import (
	"context"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// DataStorage - interface for describing storage for saving
// sensetive data.
type DataStorage interface {
	Connect() error

	LoginUser(ctx context.Context, login, password string) (string, error)

	RegisterUser(ctx context.Context, login, password, email string) (string, error)

	CheckUserJWT(ctx context.Context, userLogin string) (string, error)

	UploadPassword(ctx context.Context, password, app, md string, initVector []byte, updatedAt time.Time) error

	UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string, initVector []byte, updatedAt time.Time) error

	UploadFile(ctx context.Context, fileName, metaData string, updatedAt time.Time) (bool, error)

	DeleteFile(ctx context.Context, fileName string) error

	DeleteBankCard(ctx context.Context, cardNumber string) error

	DeletePassword(ctx context.Context, application string) error

	UpdateFile(ctx context.Context, fileName, metaData string, updatedAt time.Time) (bool, error)

	UpdateBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string, updatedAt time.Time, initVector []byte) error

	UpdatePassword(ctx context.Context, password, app, md string, updatedAt time.Time, initVector []byte) error

	GetAllPasswords(ctx context.Context, passwordInfo *[]*pb.PasswordMessage) (map[string][]byte, error)

	GetAllCardsCredentials(ctx context.Context, bankCardsInfo *[]*pb.BankCardMessage) (map[string][]byte, error)

	GetAllFilesInfo(ctx context.Context) (map[string]ut.FileInfo, error)

	GetPassword(ctx context.Context, application string) (pb.PasswordMessage, []byte, error)

	GetBankCardCredentials(ctx context.Context, cardNumber string) (*pb.BankCardMessage, []byte, error)

	GetFile(ctx context.Context, fileName string) (*pb.FileMessage, error)
}
