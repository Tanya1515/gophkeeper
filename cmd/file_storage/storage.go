package filestorage

type FileStorage interface {
	Connect()

	CreateUserFileStorage(bucketName string) error 

	GetFile(bucketName string, fileName string) error

	UploadFile(bucketName string, fileName string) error

	DeleteFile(bucketName string, fileName string) error
}
