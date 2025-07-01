// Datastorage - package, that describes interface for using
// database in the context of sensetive data storage.
package datastorage

import (
	"context"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
)

// DataStorage - interface for describing storage for saving
// sensetive data.
type DataStorage interface {
	Connect() error

	LoginUser(ctx context.Context, login, password string) (string, error)

	RegisterUser(ctx context.Context, login, password, email string) (string, error)

	CheckUserJWT(ctx context.Context, userLogin string) (string, error)

	UploadPassword(ctx context.Context, passwrod, app, md string, updatedAt time.Time, initVector []byte) error

	UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string, updatedAt time.Time, initVector []byte) error

	UploadFile(ctx context.Context, fileName, metaData string, updatedAt time.Time) error

	DeleteFile(ctx context.Context, fileName string) error

	DeleteBankCard(ctx context.Context, cardNumber string) error

	DeletePassword(ctx context.Context, application string) error

	GetAllPasswords(ctx context.Context, passwordInfo *[]*pb.PasswordMessage) (map[string][]byte, error)

	GetAllCardsCredentials(ctx context.Context, bankCardsInfo *[]*pb.BankCardMessage) (map[string][]byte, error)

	GetAllFilesInfo(ctx context.Context) (map[string]string, error)

	GetPassword(ctx context.Context, application string) (pb.PasswordMessage, []byte, error)

	GetBankCardCredentials(ctx context.Context, cardNumber string) (*pb.BankCardMessage, []byte, error)

	UpdatePassword(ctx context.Context, password, app, md string, updatedAt time.Time, initVector []byte) error

	UpdateFile(ctx context.Context, fileName, metaData string, updatedAt time.Time) error

	UpdateBankCardCreds(ctx context.Context, cardNumber, cvc, date, bank, md string, updatedAt time.Time, initVector []byte) error
}
