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

func (m *MinioStorage) Connect() {
	minioClient, err := minio.New(m.Endpoint,
		m.AccessKeyID,
		m.SecretAccessKey,
		m.UseSSL,
	)

	if err != nil {
		fmt.Println("Error while connecting to minio: ", err)
	}

	m.minioClient = minioClient
}

func (m *MinioStorage) GetFile() {

}

func (m *MinioStorage) UploadFile() {

}

func (m *MinioStorage) DeleteFile() {

}

func (m *MinioStorage) CreateUserFileStorage() {

}
