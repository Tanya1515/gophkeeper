package client

import (
	"context"
	"fmt"
	"os"

	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	"google.golang.org/grpc/metadata"
)

func ExecuteOperationFiles(operation cs.Operation, userJWT, uploadTime string, fieldName, fieldValue []string) {

	md := metadata.New(map[string]string{"Authorization": userJWT})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	certPath, envExists := os.LookupEnv("CERT_PATH")
	if !(envExists) {
		certPath = "../../test_certs/"
	}

	connection, err := ClientConnection(certPath)
	if err != nil {
		fmt.Println("Error while creating GRPC connection to server: ", err)
	}

	clientGRPC := pb.NewGophkeeperClient(connection)

	switch operation {
	case cs.Create:
	case cs.Get:
	case cs.Update:
	case cs.Delete:
	}
}

func ExecuteOperationCards(operation cs.Operation, userJWT, uploadTime string, fieldName, fieldValue []string) {

	md := metadata.New(map[string]string{"Authorization": userJWT})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	certPath, envExists := os.LookupEnv("CERT_PATH")
	if !(envExists) {
		certPath = "../../test_certs/"
	}

	connection, err := ClientConnection(certPath)
	if err != nil {
		fmt.Println("Error while creating GRPC connection to server: ", err)
	}

	clientGRPC := pb.NewGophkeeperClient(connection)

	switch operation {
	case cs.Create:
	case cs.Get:
	case cs.Update:
	case cs.Delete:
	}
}

func ExecuteOperationPasswords(operation cs.Operation, userJWT, uploadTime string, fieldName, fieldValue []string) {

	md := metadata.New(map[string]string{"Authorization": userJWT})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	certPath, envExists := os.LookupEnv("CERT_PATH")
	if !(envExists) {
		certPath = "../../test_certs/"
	}

	connection, err := ClientConnection(certPath)
	if err != nil {
		fmt.Println("Error while creating GRPC connection to server: ", err)
	}

	clientGRPC := pb.NewGophkeeperClient(connection)

	switch operation {
	case cs.Create:
	case cs.Get:
	case cs.Update:
	case cs.Delete:
	}
}

func (c *Client) SendCacheData() {
	cardOprations, err := c.ClientStorage.GetAllCardOperation()
	if err != nil {
		return
	}

	filesOperations, err := c.ClientStorage.GetAllFileOperation()
	if err != nil {
		return
	}

	passwordsOperations, err := c.ClientStorage.GetAllPasswordOperation()
	if err != nil {
		return
	}

}
