package client

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// SendPassword creates handler for processing cli command, that sends user password for application of the current user.
func (c *Client) SendPassword() *cobra.Command {
	var SendPassword = &cobra.Command{
		Use:   "password",
		Short: "Save password",
		Long:  `Save password from third-party service`,
		Run: func(cmd *cobra.Command, args []string) {
			var application, encryptedPassword, metadataPassword, password string
			var initVector []byte

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
				return
			}

			fmt.Print("Please enter password to save: ")

			password, _ = reader.ReadString('\n')
			password = strings.TrimRight(password, "\n")

			for password == "" {
				fmt.Print("Please enter password to save: ")
				password, _ = reader.ReadString('\n')
				password = strings.TrimRight(password, "\n")
			}

			fmt.Print("Please enter appplication, that password belongs to: ")

			application, _ = reader.ReadString('\n')
			application = strings.TrimRight(application, "\n")

			for application == "" {
				fmt.Print("Please enter appplication, that password belongs to: ")
				application, _ = reader.ReadString('\n')
				application = strings.TrimRight(application, "\n")
			}

			encryptedPassword, initVector, err = c.EncryptData(password)
			if err != nil {
				c.ClientLogger.Errorln("Error while encrypt sensetive data for application %s: %s", application, err)
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				return
			}

			fmt.Print("Please enter metadata for sensetive data: ")
			metadataPassword, _ = reader.ReadString('\n')
			metadataPassword = strings.TrimRight(metadataPassword, "\n")

			fmt.Println("Process your application sensetive data...")
			certPath, envExists := os.LookupEnv("CERT_PATH")
			if !(envExists) {
				certPath = "../../test_certs/"
			}

			connection, err := c.ClientConnection(certPath)
			if err != nil {
				c.ClientLogger.Errorln("Error while creating GRPC connection to server: ", err)
			}

			clientGRPC := pb.NewGophkeeperClient(connection)
			md := metadata.New(map[string]string{"Authorization": JWTToken})

			ctx := metadata.NewOutgoingContext(context.Background(), md)

			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)
			var retryCount = 1
			_, err = clientGRPC.UploadPassword(ctx, &pb.PasswordMessage{
				Password:    password,
				Application: application,
				MetaData:    metadataPassword,
				UploadTime:  opTime,
			})

			for err != nil && retryCount != 3 {
				_, err = clientGRPC.UploadPassword(ctx, &pb.PasswordMessage{
					Password:    password,
					Application: application,
					MetaData:    metadataPassword,
					UploadTime:  opTime,
				})
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}

			errUpload := c.ClientStorage.UploadPassword(application, encryptedPassword, metadataPassword, opTime, User, initVector)
			if errUpload != nil {
				c.ClientLogger.Errorln("Error while writting password data to SQLite: ", errUpload)
			}
			if err == nil {
				c.SendCacheData()
				fmt.Printf("Your password for application %s has been successfully uploaded!\n", application)
			} else {
				err = c.ClientStorage.SavePasswordOperation(application, User, cs.Create, nil, opTime)
				if err != nil {
					c.ClientLogger.Errorln("Error while saving info about password operation: ", err)
				}
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
			}

		},
	}

	return SendPassword

}

// GetPassword creates handler for processing cli command, that gets user password for the application of the current user.
func (c *Client) GetPassword() *cobra.Command {
	var GetPassword = &cobra.Command{
		Use:   "password",
		Short: "Get password of the application from gophkeeper",
		Run: func(cmd *cobra.Command, args []string) {
			var application string

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
			}

			fmt.Print("Please enter appplication, that password belongs to: ")
			application, _ = reader.ReadString('\n')
			application = strings.TrimRight(application, "\n")

			for application == "" {
				fmt.Print("Please enter appplication, that password belongs to: ")
				application, _ = reader.ReadString('\n')
				application = strings.TrimRight(application, "\n")
			}
			fmt.Println("Process your application sensetive data...")
			c.SendCacheData()
			password, uploadTime, metadataPassword, initVector, exists, err := c.ClientStorage.GetPassword(application, User)
			if exists {
				passwordDecrypted, err := c.Crypto.DecryptData(password, initVector)
				if err != nil {
					c.ClientLogger.Errorln("Error while decrypting sensetive data for application %s: %s", application, err)
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
					return
				}
				fmt.Printf("Application: %s\n", application)
				fmt.Printf("Password: %s\n", passwordDecrypted)
				fmt.Printf("Upload date: %s\n", uploadTime)
				fmt.Printf("Additioanl information: %s\n", metadataPassword)
			} else {
				if err != nil {
					c.ClientLogger.Errorf("Error while getting application %s credentials from local storage for user %s: %w", application, User, err)
				}
				certPath, envExists := os.LookupEnv("CERT_PATH")
				if !(envExists) {
					certPath = "../../test_certs/"
				}

				connection, err := c.ClientConnection(certPath)
				if err != nil {
					c.ClientLogger.Errorln("Error while creating GRPC connection to server: ", err)
				}

				clientGRPC := pb.NewGophkeeperClient(connection)
				md := metadata.New(map[string]string{"Authorization": JWTToken})

				ctx := metadata.NewOutgoingContext(context.Background(), md)

				var retryCount = 1
				passwordApp, err := clientGRPC.GetPassword(ctx, &pb.SensetiveDataMessage{
					Identificator: application,
				})
				if err != nil {
					c.ClientLogger.Errorf("Error while getting password for application %s: %s\n", application, err)
				}
				for err != nil && retryCount != 3 {
					passwordApp, err = clientGRPC.GetPassword(ctx, &pb.SensetiveDataMessage{
						Identificator: application,
					})
					if err != nil {
						c.ClientLogger.Errorf("Error while getting password for application %s: %s\n", application, err)
					}
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}

				if err == nil {
					fmt.Printf("Application: %s\n", application)
					fmt.Printf("Password: %s\n", passwordApp.Password)
					fmt.Printf("Additioanl information: %s\n", passwordApp.MetaData)
				} else if strings.Contains(strings.Split(err.Error(), "desc")[1], "no rows in result") {
					fmt.Printf("Sensetive data for application %s do not exist\n", application)
				} else {
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				}
			}
		},
	}

	return GetPassword

}

// DeletePassword creates handler for processing cli command, that deletes user password for the application of the current user.
func (c *Client) DeletePassword() *cobra.Command {

	var DeletePassword = &cobra.Command{
		Use:   "password",
		Short: "Delete password of the application from gophkeeper",
		Run: func(cmd *cobra.Command, args []string) {
			var application string

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
			}

			fmt.Print("Please enter appplication, that password belongs to: ")
			application, _ = reader.ReadString('\n')
			application = strings.TrimRight(application, "\n")

			for application == "" {
				fmt.Print("Please enter appplication, that password belongs to: ")
				application, _ = reader.ReadString('\n')
				application = strings.TrimRight(application, "\n")
			}

			fmt.Println("Process your application sensetive data...")
			c.SendCacheData()

			certPath, envExists := os.LookupEnv("CERT_PATH")
			if !(envExists) {
				certPath = "../../test_certs/"
			}

			connection, err := c.ClientConnection(certPath)
			if err != nil {
				c.ClientLogger.Errorln("Error while creating GRPC connection to server: ", err)
			}

			clientGRPC := pb.NewGophkeeperClient(connection)
			md := metadata.New(map[string]string{"Authorization": JWTToken})

			ctx := metadata.NewOutgoingContext(context.Background(), md)

			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)

			var retryCount = 1
			_, err = clientGRPC.DeletePassword(ctx, &pb.SensetiveDataMessage{
				Identificator: application,
			})
			if err != nil {
				c.ClientLogger.Errorf("Error while deleting password in Gophkeeper for application %s: %s\n", application, err)
			}

			for err != nil && retryCount != 3 {
				_, err = clientGRPC.DeletePassword(ctx, &pb.SensetiveDataMessage{
					Identificator: application,
				})
				if err != nil {
					c.ClientLogger.Errorf("Error while deleting password in Gophkeeper for application %s: %s\n", application, err)
				}
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}

			errLocal := c.ClientStorage.DeletePassword(application, User)
			if errLocal != nil {
				c.ClientLogger.Errorf("Error while deleting password for application %s from local storage: %s \n", application, err)
			}
			if err == nil || errors.Is(err, sql.ErrNoRows) {
				fmt.Printf("All sensetive data regarding to application %s was successfully removed from gophkeeper", application)
			} else {
				err = c.ClientStorage.SavePasswordOperation(application, User, cs.Delete, nil, opTime)
				if err != nil {
					c.ClientLogger.Errorln("Error while saving info about password operation: ", err)
				}
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
			}
		},
	}

	return DeletePassword

}

// UpdatePassword creates handler for processing cli command, that updates user password for the application of the current user.
func (c *Client) UpdatePassword() *cobra.Command {
	var UpdatePassword = &cobra.Command{
		Use:   "password",
		Short: "Update password of the application from gophkeeper",
		Run: func(cmd *cobra.Command, args []string) {
			var newPassword, passwordMetadata, application, encryptedPassword string
			var initVector []byte

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
			}

			fmt.Print("Please enter appplication, that password belongs to: ")
			application, _ = reader.ReadString('\n')
			for application == "" {
				fmt.Print("Please enter appplication, that password belongs to: ")
				application, _ = reader.ReadString('\n')
			}

			fmt.Println("Process your application sensetive data...")
			c.SendCacheData()

			application = strings.TrimRight(application, "\n")

			fmt.Print("Please enter new password: ")
			newPassword, _ = reader.ReadString('\n')
			newPassword = strings.TrimRight(newPassword, "\n")

			fmt.Print("Please enter metadata: ")
			passwordMetadata, _ = reader.ReadString('\n')
			passwordMetadata = strings.TrimRight(passwordMetadata, "\n")

			for newPassword == "" && passwordMetadata == "" {
				fmt.Printf("Please enter password or metadata for application %s for updating", application)
				fmt.Println("Please enter new password: ")
				newPassword, _ = reader.ReadString('\n')

				fmt.Println("Please enter metadata: ")
				passwordMetadata, _ = reader.ReadString('\n')
			}

			if newPassword != "" {
				encryptedPassword, initVector, err = c.EncryptData(newPassword)
				if err != nil {
					c.ClientLogger.Errorln("Error while encrypt sensetive data for application %s: %s", application, err)
					return
				}
			}

			certPath, envExists := os.LookupEnv("CERT_PATH")
			if !(envExists) {
				certPath = "../../test_certs/"
			}

			connection, err := c.ClientConnection(certPath)
			if err != nil {
				c.ClientLogger.Infoln("Error while creating GRPC connection to server: ", err)
			}

			clientGRPC := pb.NewGophkeeperClient(connection)
			md := metadata.New(map[string]string{"Authorization": JWTToken})

			ctx := metadata.NewOutgoingContext(context.Background(), md)

			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)

			var retryCount = 1
			_, err = clientGRPC.UpdatePassword(ctx, &pb.PasswordMessage{
				Password:    newPassword,
				Application: application,
				MetaData:    passwordMetadata,
				UploadTime:  opTime,
			})
			if err != nil {
				c.ClientLogger.Errorf("Error while updating password for application %s: %s\n", application, err)
			}

			for err != nil && retryCount != 3 {
				_, err = clientGRPC.UpdatePassword(ctx, &pb.PasswordMessage{
					Password:    newPassword,
					Application: application,
					MetaData:    passwordMetadata,
					UploadTime:  opTime,
				})
				if err != nil {
					c.ClientLogger.Errorf("Error while updating password for application %s: %s\n", application, err)
				}
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}
			uploadErr := c.ClientStorage.UploadPassword(application, encryptedPassword, passwordMetadata, opTime, User, initVector)
			if uploadErr != nil {
				c.ClientLogger.Errorf("Error while updating  password sensetive data for application %s: %s\n", application, err)
			}
			if err == nil {
				fmt.Printf("Your password for application %s has been successfully updated!\n", application)
			} else {
				fields := make([]string, 0)
				if newPassword != "" {
					fields = append(fields, "password")
				}
				if passwordMetadata != "" {
					fields = append(fields, "metadata")
				}
				err = c.ClientStorage.SavePasswordOperation(application, User, cs.Update, fields, opTime)
				if err != nil {
					c.ClientLogger.Errorf("Error while saving information about update operation for sensetive data of application %s: %s\n", application, err)
				}
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
			}

		},
	}

	return UpdatePassword
}

// ExecutePasswordsOperation - function for executing old operations with sensetive data for application.
func (c *Client) ExecutePasswordsOperation(operation cs.Operation, userJWT string, password *pb.PasswordMessage) error {

	md := metadata.New(map[string]string{"Authorization": userJWT})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	certPath, envExists := os.LookupEnv("CERT_PATH")
	if !(envExists) {
		certPath = "../../test_certs/"
	}

	connection, err := c.ClientConnection(certPath)
	if err != nil {
		return fmt.Errorf("error while creating GRPC connection to server: %w", err)
	}

	clientGRPC := pb.NewGophkeeperClient(connection)

	switch operation {
	case cs.Create:
		_, err = clientGRPC.UploadPassword(ctx, password)
		if err != nil {
			return fmt.Errorf("error while uploading password: %w", err)
		}
	case cs.Get:
		password, err = clientGRPC.GetPassword(ctx, &pb.SensetiveDataMessage{
			Identificator: password.Application,
		})
		if err != nil {
			return fmt.Errorf("error while getting password: %w", err)
		}
	case cs.Update:
		_, err = clientGRPC.UpdatePassword(ctx, password)
		if err != nil {
			return fmt.Errorf("error while updating password: %w", err)
		}
	case cs.Delete:
		_, err = clientGRPC.DeletePassword(ctx, &pb.SensetiveDataMessage{
			Identificator: password.Application,
		})
		if err != nil {
			return fmt.Errorf("error while deleting password: %w", err)
		}
	}

	return nil
}
