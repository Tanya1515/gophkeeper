package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login into gophkeeper",
	Long:  `Command to authentificate with user login, password and OTP into gophkeeper`,
	Run: func(cmd *cobra.Command, args []string) {
		var login string
		var password string
		

		fmt.Print("Login: ")
		fmt.Fscan(os.Stdin, &login)
		fmt.Print("Password: ")
		fmt.Fscan(os.Stdin, &password)

		connection, err := ClientConnection()
		if err != nil {
			fmt.Println("Error while creating GRPC connection to server: ", err)
		}

		clientGRPC := pb.NewGophkeeperClient(connection)
		status, err := clientGRPC.LoginUser(context.Background(), &pb.User{
			Login:    login,
			Password: password,
		})

		if err != nil {
			fmt.Println("Error while sending request to grpc server ", err)
		}

		fmt.Println(status)
		defer connection.Close()
	},
}
