package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login into gophkeeper",
	Long:  `Command to authentificate with user login, password and OTP into gophkeeper`,
	Run: func(cmd *cobra.Command, args []string) {
		var login string
		var password string
		var oneTimePassword string

		fmt.Print("Login: ")
		fmt.Fscan(os.Stdin, &login)
		fmt.Print("Password: ")
		fmt.Fscan(os.Stdin, &password)

		connection, err := ClientConnection()
		if err != nil {
			fmt.Println("Error while creating GRPC connection to server: ", err)
		}

		clientGRPC := pb.NewGophkeeperClient(connection)
		_, err = clientGRPC.LoginUser(context.Background(), &pb.User{
			Login:    login,
			Password: password,
		})

		if err != nil {
			fmt.Println("Error while sending request to grpc server ", err)
		}

		fmt.Print("Please enter one-time password: ")
		fmt.Fscan(os.Stdin, &oneTimePassword)
		result, err := clientGRPC.VerificationApprove(context.Background(), &pb.Verify{
			Login:       login,
			OneTimePass: oneTimePassword,
		})

		if err != nil {
			fmt.Println("Error while checking if OTP is correct")
		}

		err = ut.SaveJWT(result.JWTtoken, login)
		if err != nil {
			fmt.Printf("Error while saving user %s JWTToken %s", login, err)
		}

		defer connection.Close()
	},
}
