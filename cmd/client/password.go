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

var sendPassword = &cobra.Command{
	Use:   "password",
	Short: "Save password",
	Long:  `Save password from third-party service`,
	Run: func(cmd *cobra.Command, args []string) {
		var password string
		var application string
		var metadataPassword string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter password to save: ")
		fmt.Fscan(os.Stdin, &password)
		fmt.Print("Please enter appplication, that password belongs to: ")
		fmt.Fscan(os.Stdin, &application)
		fmt.Print("Please enter metadata for sensetive data: ")
		fmt.Fscan(os.Stdin, &metadataPassword)

		connection, err := ClientConnection()
		if err != nil {
			fmt.Println("Error while creating GRPC connection to server: ", err)
		}

		clientGRPC := pb.NewGophkeeperClient(connection)
		md := metadata.New(map[string]string{"Authorization": JWTToken})

		ctx := metadata.NewOutgoingContext(context.Background(), md)

		_, err = clientGRPC.UploadPassword(ctx, &pb.UploadPasswordMessage{
			Password:    password,
			Application: application,
			MetaData:    metadataPassword,
		})

		if err != nil {
			fmt.Printf("Error while uploading password for application %s : %s\n", application, err)
			return
		}
		fmt.Printf("Your password for application %s has been successfully uploaded!\n", application)
	},
}

// доделать 
var getPassword = &cobra.Command{
	Use:   "password",
	Short: "Get password of the application from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var deletePassword = &cobra.Command{
	Use:   "password",
	Short: "Delete password of the application from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

var updatePassword = &cobra.Command{
	Use:   "password",
	Short: "Update password of the application from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
