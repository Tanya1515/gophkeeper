package minio

import (
	"fmt"

	"github.com/minio/minio-go"
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

func (m *MinioStorage) GetFile(bucketName string, fileName string) error {

	return nil
}

func (m *MinioStorage) UploadFile(bucketName string, fileName string) error {
	return nil
}

func (m *MinioStorage) DeleteFile(bucketName string, fileName string) error {
	return nil
}

func (m *MinioStorage) CreateUserFileStorage(bucketName string) (err error) {

	err = m.minioClient.MakeBucket(bucketName, "us-east-1")
	if err != nil {
		return fmt.Errorf("error, while creating bucket with name %s: %w", bucketName, err)
	}

	return nil
}
