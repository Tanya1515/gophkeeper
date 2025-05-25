package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

var sendFile = &cobra.Command{
	Use:   "file",
	Short: "Save file",
	Long:  `Save file with sensetive data to gophkeeper!`,
	Run: func(cmd *cobra.Command, args []string) {
		var filePath string
		var metadataFile string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter absolute path for file to save: ")
		fmt.Fscan(os.Stdin, &filePath)
		fmt.Print("Please enter metadata for sensetive data: ")
		fmt.Fscan(os.Stdin, &metadataFile)

		connection, err := ClientConnection()
		if err != nil {
			fmt.Println("Error while creating GRPC connection to server: ", err)
		}

		_ = pb.NewGophkeeperClient(connection)
		md := metadata.New(map[string]string{"Authorization": JWTToken})

		_ = metadata.NewOutgoingContext(context.Background(), md)
		// дописать отправку файла

	},
}

//доделать
var getFile = &cobra.Command{
	Use:   "file",
	Short: "Get description of all user sensetive data",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var updateFile = &cobra.Command{
	Use:   "file",
	Short: "Update existing file with sensetive data in gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var deleteFile = &cobra.Command{
	Use:   "file",
	Short: "Delete existing file with sensetive data in gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
