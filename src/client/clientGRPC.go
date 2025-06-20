package client

import (
	"fmt"
	"path/filepath"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func ClientConnection(certPath string) (*grpc.ClientConn, error) {
	certFile, err := filepath.Abs(certPath + "ca.crt")
	if err != nil {
		fmt.Println("Error while searching for ca.crt ", err)
		return nil, err
	}

	credsTLS, err := credentials.NewClientTLSFromFile(certFile, "")
	if err != nil {
		fmt.Println("Error while getting ca.cert ", err)
		return nil, err
	}

	conn, err := grpc.NewClient("localhost:3200", grpc.WithTransportCredentials(credsTLS))
	if err != nil {
		fmt.Println("Error while openning connection to server: ", err)
		return nil, err
	}

	return conn, err
}
