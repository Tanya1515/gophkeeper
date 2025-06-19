// Datastorage - package, that describes interface for using
// any file stoarge in the context of sensetive data storage.
package filestorage

import "context"

type FileStorage interface {
	Connect() error

	CreateUserFileStorage(ctx context.Context, bucketName string) error

	GetFile(ctx context.Context, fileName string) ([]byte, error)

	UploadFile(ctx context.Context, fileName, absolutePath string) error

	DeleteFile(ctx context.Context, fileName string) error
}
