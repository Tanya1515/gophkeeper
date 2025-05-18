package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register into gophkeeper",
	Long:  `Register new user with login, password in gophkeeper`,
	Run: func(cmd *cobra.Command, args []string) {
		var login string
		var password string
		var email string

		fmt.Print("Login: ")
		fmt.Fscan(os.Stdin, &login)
		fmt.Print("Password: ")
		fmt.Fscan(os.Stdin, &password)
		fmt.Print("User's email: ")
		fmt.Fscan(os.Stdin, &email)

		connection, err := ClientConnection()
		if err != nil {
			fmt.Println("Error while creating GRPC connection to server: ", err)
		}

		clientGRPC := pb.NewGophkeeperClient(connection)
		status, err := clientGRPC.RegisterUser(context.Background(), &pb.User{
			Login:    login,
			Password: password,
			Email:    email,
		})

		if err != nil {
			fmt.Println("Error while sending request to grpc server ", err)
		}

		fmt.Println(status)
		defer connection.Close()
	},
}
