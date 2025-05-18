package datastorage

type DataStorage interface {
	Connect() error

	LoginUser() error

	RegisterUser() error
}
