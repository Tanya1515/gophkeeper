// Minio - package, that containts functions for managing user files
// in Minio buckets, including delete, update, upload and get operations.
package minio

import (
	"context"
	"fmt"
	"io"

	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
	"github.com/minio/minio-go"
)

type MinioStorage struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	minioClient     *minio.Client
}

// NewMinioStorage - function for setting Minio credentials.
func NewMinioStorage(endpoint, accessKeyID, secretAccessKey string, useSSL bool) *MinioStorage {
	minioStoreClient := MinioStorage{Endpoint: endpoint, AccessKeyID: accessKeyID, SecretAccessKey: secretAccessKey, UseSSL: useSSL}
	return &minioStoreClient
}

// Connect - function for connecting to Minio instance.
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

// GetFile - function for getting file from Minio bucket for current user.
func (m *MinioStorage) GetFile(ctx context.Context, fileName string) ([]byte, error) {

	minioFile, err := m.minioClient.GetObject(ctx.Value(ut.LoginKey).(string), fileName, minio.GetObjectOptions{})
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error while getting file %s from Minio: %w", fileName, err)
	}

	fileInfo, err := minioFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("error while getting file %s info from Minio object: %w", fileName, err)
	}

	resFile := make([]byte, fileInfo.Size)

	_, err = minioFile.Read(resFile)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error while reading file %s from Minio: %w", fileName, err)
	}

	return resFile, nil
}

// UploadFile - function for uploading new file to user Minio bucket.
func (m *MinioStorage) UploadFile(ctx context.Context, fileName, absolutePath string) error {
	_, err := m.minioClient.FPutObject(ctx.Value(ut.LoginKey).(string), fileName, absolutePath, minio.PutObjectOptions{})
	if err != nil {
		return err
	}

	return nil
}

// DeleteFile - function for deleting file from user bucket in Minio.
func (m *MinioStorage) DeleteFile(ctx context.Context, fileName string) (err error) {
	err = m.minioClient.RemoveObject(ctx.Value(ut.LoginKey).(string), fileName)

	return
}

// CreateUserFileStorage - function for creating bucket in Minio (for hosting files of the current user).
func (m *MinioStorage) CreateUserFileStorage(ctx context.Context, bucketName string) (err error) {

	err = m.minioClient.MakeBucket(bucketName, "us-east-1")
	if err != nil {
		return fmt.Errorf("error, while creating bucket with name %s: %w", bucketName, err)
	}

	return nil
}
