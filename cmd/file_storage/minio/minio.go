package minio

import (
	"context"
	"fmt"

	"github.com/minio/minio-go"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

type MinioStorage struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	minioClient     *minio.Client
}

func NewMinioStorage(endpoint, accessKeyID, secretAccessKey string, useSSL bool) *MinioStorage {
	minioStoreClient := MinioStorage{Endpoint: endpoint, AccessKeyID: accessKeyID, SecretAccessKey: secretAccessKey, UseSSL: useSSL}
	return &minioStoreClient
}

func (m *MinioStorage) Connect() (err error) {
	minioClient, err := minio.New(m.Endpoint,
		m.AccessKeyID,
		m.SecretAccessKey,
		m.UseSSL,
	)

	if err != nil {
		return
	}
	m.minioClient = minioClient

	return
}

func (m *MinioStorage) GetFile(ctx context.Context, fileName string) error {

	return nil
}

func (m *MinioStorage) UploadFile(ctx context.Context, fileName, absolutePath string) error {
	_, err := m.minioClient.FPutObject(ctx.Value(ut.LoginKey).(string), fileName, absolutePath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (m *MinioStorage) DeleteFile(ctx context.Context, fileName string) (err error) {
	err = m.minioClient.RemoveObject(ctx.Value(ut.LoginKey).(string), fileName)

	return
}

func (m *MinioStorage) CreateUserFileStorage(ctx context.Context, bucketName string) (err error) {

	err = m.minioClient.MakeBucket(bucketName, "us-east-1")
	if err != nil {
		return fmt.Errorf("error, while creating bucket with name %s: %w", bucketName, err)
	}

	return nil
}
