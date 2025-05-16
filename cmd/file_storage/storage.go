package filestorage

type FileStorage interface {
	Connect()
	CreateUserFileStorage()
	GetFile()
	UploadFile()
	DeleteFile()
}
