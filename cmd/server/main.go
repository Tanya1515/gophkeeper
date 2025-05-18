package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	dataStorage "github.com/Tanya1515/gophkeeper.git/cmd/data_storage"
	postgresql "github.com/Tanya1515/gophkeeper.git/cmd/data_storage/postgresql"
	fileStorage "github.com/Tanya1515/gophkeeper.git/cmd/file_storage"
	minio "github.com/Tanya1515/gophkeeper.git/cmd/file_storage/minio"
	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

type GophkeeperServer struct {
	
	DataStorage dataStorage.DataStorage// DataStorage saves all user sensetive data

	FileStorage fileStorage.FileStorage// FileStorage saves all user files 

	Logger zap.SugaredLogger // Logger saves all server info
	
	pb.UnimplementedGophkeeperServer // type pb.Unimplemented<TypeName> is used for backward compatibility
}

func generateTLSCreds() (credentials.TransportCredentials, error) {
	certFile, err := filepath.Abs("../../certs/server.crt")
	if err != nil {
		fmt.Println("Error while searching for server.crt ", err)
		return nil, err
	}
	keyFile, err := filepath.Abs("../../certs/server.key")
	if err != nil {
		fmt.Println("Error while searching for server.key ", err)
		return nil, err
	}

	return credentials.NewServerTLSFromFile(certFile, keyFile)
}

func main() {
	var s *grpc.Server
	address := "localhost:3200"

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	defer logger.Sync()

	loggerApp := *logger.Sugar()

	endpoint, envExists := os.LookupEnv("MINIO_HOST")
	if !(envExists) {
		loggerApp.Errorln("Error while getting postgreSQL address")
		return
	}
	
	accessKeyID, envExists := os.LookupEnv("MINIO_ROOT_USER")
	if !(envExists) {
		loggerApp.Errorln("Error while getting access key id for Minio")
		return
	}

	secretAccessKey, envExists := os.LookupEnv("MINIO_ROOT_PASSWORD")
	if !(envExists) {
		loggerApp.Errorln("Error while getting secret access key for Minio")
		return
	}

	minioStorage := minio.NewMinioStorage(endpoint, accessKeyID, secretAccessKey, true)

	host, envExists := os.LookupEnv("POSTGRES_HOST")
	if !(envExists) {
		loggerApp.Errorln("Error while getting address for PostgreSQL")
		return
	}

	userName, envExists := os.LookupEnv("POSTGRES_USER")
	if !(envExists) {
		loggerApp.Errorln("Error while getting userName for PostgreSQL")
		return
	}

	password, envExists := os.LookupEnv("POSTGRES_PASSWORD")
	if !(envExists) {
		loggerApp.Errorln("Error while getting password for PostgreSQL")
		return
	}

	dbName, envExists := os.LookupEnv("POSTGRES_DB")
	if !(envExists) {
		loggerApp.Errorln("Error while getting database name for PostgreSQL")
		return
	}

	postgreSQL := postgresql.NewPostgreSQLConnection(host, userName, password, dbName)

	listen, err := net.Listen("tcp", address)
	if err != nil {
		loggerApp.Errorln("Error while openning connection on address ", address, " : ", err)
	}
	credsTLS, err := generateTLSCreds()
	if err != nil {
		loggerApp.Errorln("Error while getting certificates for GRPC server ", err)
	}
	s = grpc.NewServer(grpc.Creds(credsTLS))
	

	gophkeeper := &GophkeeperServer{Logger: loggerApp, DataStorage: postgreSQL, FileStorage: minioStorage}
	pb.RegisterGophkeeperServer(s, gophkeeper)
	if err := s.Serve(listen); err != nil {
		fmt.Println("Error, while trying to start grpc server: ", err)
	}
}
