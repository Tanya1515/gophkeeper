package datastorage

import (
	"context"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

type DataStorage interface {
	Connect() error

	LoginUser(ctx context.Context, login, password string) (string, error)

	RegisterUser(ctx context.Context, login, password, email string) (string, error)

	CheckUserJWT(ctx context.Context, userLogin string) (string, error)

	UploadPassword(ctx context.Context, passwrod, app, md string) error

	UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string) error

	UploadFile(ctx context.Context, fileName, metaData string) error

	DeleteFile(ctx context.Context, fileName string) error

	DeleteBankCard(ctx context.Context, cardNumber string) error

	DeletePassword(ctx context.Context, application string) error

	GetPassword(ctx context.Context, application string) (pb.PasswordMessage, error)

	GetBankCardCredentials(ctx context.Context, cardNumber string) (*pb.BankCardMessage, error)
}
