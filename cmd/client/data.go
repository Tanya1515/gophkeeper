package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func SyncAllFiles(wg *sync.WaitGroup, JWTToken string, clientGRPC pb.GophkeeperClient, files map[string]string) {
	var fileName string
	var fileToSave *os.File
	md := metadata.New(map[string]string{"Authorization": JWTToken})

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := clientGRPC.SyncFiles(ctx, &emptypb.Empty{})
	if err != nil {
		fmt.Printf("Error while grpc connection set up: %s\n", err)
		return
	}

	file, err := stream.Recv()
	if err != nil && err != io.EOF {
		fmt.Printf("Error while recieving file chunk %s from gophkeeper: %s\n", file.FileName, err)
		return
	}

	if fileName != file.FileName {
		fileName = file.FileName
		fileToSave, err = os.Create("/tmp/" + fileName)
		if err != nil {
			fmt.Printf("Error while creating file with path %s: %s\n", "/tmp/"+fileName, err)
			return
		}
	}

	_, err = fileToSave.Write(file.Content)
	if err != nil {
		fmt.Printf("Error while saving file %s: %s\n", fileName, err)
		return
	}

	for {
		file, err := stream.Recv()
		if err != nil && err != io.EOF {
			fmt.Printf("Error while recieving file chunk %s from gophkeeper: %s\n", file.FileName, err)
			return
		}
		if file.End {
			files[fileName] = "/tmp/" + fileName
			break
		}

		if fileName != file.FileName {
			files[fileName] = "/tmp/" + fileName
			fileName = file.FileName
			fileToSave, err = os.Create("/tmp/" + fileName)
			if err != nil {
				fmt.Printf("Error while creating file with path %s: %s\n", "/tmp/"+fileName, err)
				return
			}
		}

		_, err = fileToSave.Write(file.Content)
		if err != nil {
			fmt.Printf("Error while saving file %s: %s\n", fileName, err)
			return
		}

	}
	defer wg.Done()
}

var getUserData = &cobra.Command{
	Use:   "all",
	Short: "Get description of all user sensetive data",
	Run: func(cmd *cobra.Command, args []string) {
		var wg sync.WaitGroup
		files := make(map[string]string, 100)
		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
			return
		}

		certPath, envExists := os.LookupEnv("CERT_PATH")
		if !(envExists) {
			certPath = "../../test_certs/"
		}

		connection, err := ClientConnection(certPath)
		if err != nil {
			fmt.Println("Error while creating GRPC connection to server: ", err)
		}

		clientGRPC := pb.NewGophkeeperClient(connection)
		md := metadata.New(map[string]string{"Authorization": JWTToken})

		ctx := metadata.NewOutgoingContext(context.Background(), md)

		wg.Add(1)

		go SyncAllFiles(&wg, JWTToken, clientGRPC, files)

		sensetiveData, err := clientGRPC.Sync(ctx, &emptypb.Empty{})
		if err != nil {
			fmt.Printf("Error while getting all sensetive data for user %s: %s\n", user, err)
			return
		}
		fmt.Println("Yor passwords: ")
		for _, passwordInfo := range sensetiveData.Passwords {
			fmt.Printf("\n  Application: %s,\n   password: %s,\n   Metadata: %s\n", passwordInfo.Application, passwordInfo.Password, passwordInfo.MetaData)
		}

		fmt.Println("\nYour bank card credentials: ")
		fmt.Println()
		for _, bankCardCreds := range sensetiveData.BankCards {
			fmt.Println("  Bank: ", bankCardCreds.Bank)
			fmt.Println("    Card number: ", bankCardCreds.CardNumber)
			fmt.Printf("    CVC: %s, date: %s\n", bankCardCreds.CvcCode, bankCardCreds.Data)
			fmt.Printf("    Metada: %s\n\n", bankCardCreds.Metadata)
		}

		wg.Wait()
		if len(files) != 0 {
			fmt.Println("Your files: ")
			for fileName, pathToFile := range files {
				fmt.Printf("\n  File %s have been saved along the path: %s\n", fileName, pathToFile)
			}
			fmt.Println()
		}

		fmt.Printf("All data for user %s was synchronyzed\n", user)

	},
}
