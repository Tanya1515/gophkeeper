package client

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// RegisterClient creates handler for processing cli command, that registers user in Gophkeeper.
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

			c.ClientLogger.Infof("Start processing user %s", User)
			certPath, envExists := os.LookupEnv("CERT_PATH")
			if !(envExists) {
				certPath = "../../test_certs/"
			}

			connection, err := c.ClientConnection(certPath)
			if err != nil {
				c.ClientLogger.Errorln("Error while creating GRPC connection to server: ", err)
			}

			clientGRPC := pb.NewGophkeeperClient(connection)
			_, err = clientGRPC.RegisterUser(context.Background(), &pb.User{
				Login:    login,
				Password: password,
				Email:    email,
			})

			if err != nil {
				c.ClientLogger.Errorln("Error while sending request to grpc server ", err)
				return
			}

			fmt.Print("Please enter one-time password: ")
			fmt.Fscan(os.Stdin, &oneTimePassword)
			result, err := clientGRPC.VerificationApprove(context.Background(), &pb.Verify{
				Login:       login,
				OneTimePass: oneTimePassword,
			})

			if err != nil {
				c.ClientLogger.Errorln("Error while checking if OTP is correct")
				return
			}

			err = ut.SaveJWT(result.JWTtoken, login)
			if err != nil {
				c.ClientLogger.Errorf("Error while saving user %s JWTToken %s", login, err)
				return
			}

			fmt.Println(login, " has been successfully registered!")

			err = c.ClientStorage.CreateUser(login)
			if err != nil {
				c.ClientLogger.Errorln(err)
			}

			c.ClientLogger.Infof("Successfully registered user %s", User)
			defer connection.Close()
		},
	}

	return RegisterCmd

}
