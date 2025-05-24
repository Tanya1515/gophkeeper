package datastorage

import "context"

type DataStorage interface {
	Connect() error

	LoginUser(ctx context.Context, login, password string) (string, error)

	RegisterUser(ctx context.Context, login, password, email string) error

	CheckUserJWT(ctx context.Context, userLogin string) (string, error)

	UploadPassword(ctx context.Context, passwrod, app, md string) (error)

	UploadBankCard(ctx context.Context, cardNumber, cvc, date, bank, md string) (error)
}
