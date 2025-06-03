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
