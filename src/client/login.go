// Client - package for running console client for uploading
// such sensetive data as passwords, files and bank card
// credentials.
package client

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// RegisterClient creates handler for processing cli command, that logins user in Gophkeeper.
func (c *Client) LoginClient() *cobra.Command {
	var LoginCmd = &cobra.Command{
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

			certPath, envExists := os.LookupEnv("CERT_PATH")
			if !(envExists) {
				certPath = "../../test_certs/"
			}

			connection, err := c.ClientConnection(certPath)
			if err != nil {
				c.ClientLogger.Errorln("Error while creating GRPC connection to server: ", err)
			}

			clientGRPC := pb.NewGophkeeperClient(connection)
			_, err = clientGRPC.LoginUser(context.Background(), &pb.User{
				Login:    login,
				Password: password,
			})

			if err != nil {
				c.ClientLogger.Errorln("Error while sending request to grpc server ", err)
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

			err = c.ClientStorage.CreateUser(login)
			if err != nil {
				c.ClientLogger.Errorln(err)
				return
			}

			defer connection.Close()
		},
	}

	return LoginCmd
}
