package client

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

func (c *Client) RegisterClient() *cobra.Command {
	var RegisterCmd = &cobra.Command{
		Use:   "register",
		Short: "Register into gophkeeper",
		Long:  `Register new user with login, password in gophkeeper`,
		Run: func(cmd *cobra.Command, args []string) {
			var login string
			var password string
			var email string
			var oneTimePassword string

			fmt.Print("Login: ")
			fmt.Fscan(os.Stdin, &login)
			fmt.Print("Password: ")
			fmt.Fscan(os.Stdin, &password)
			fmt.Print("User's email: ")
			fmt.Fscan(os.Stdin, &email)

			certPath, envExists := os.LookupEnv("CERT_PATH")
			if !(envExists) {
				certPath = "../../test_certs/"
			}

			connection, err := ClientConnection(certPath)
			if err != nil {
				fmt.Println("Error while creating GRPC connection to server: ", err)
			}

			clientGRPC := pb.NewGophkeeperClient(connection)
			_, err = clientGRPC.RegisterUser(context.Background(), &pb.User{
				Login:    login,
				Password: password,
				Email:    email,
			})

			if err != nil {
				fmt.Println("Error while sending request to grpc server ", err)
				return
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

			fmt.Println(login, " has been successfully registered!")
			defer connection.Close()
		},
	}

	return RegisterCmd

}
