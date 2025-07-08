package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

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
				fmt.Println("Please login or register to Gophkeeper.")
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
				c.ClientLogger.Errorf("Error while encrypt sensetive data for application %s: %s\n", application, err)
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

			c.SendCacheData()

			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					c.SyncData(User)
					err = c.ClientStorage.UpdateCacheMiss(User, 0)
					if err != nil {
						c.ClientLogger.Errorln(err)
					}
				}
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
			if err != nil {
				c.ClientLogger.Errorf("Error while sending sensetive data for application %s: %s\n", application, err)
				if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
					fmt.Println("Please login to Gophkeeper!")
					return
				}
			}

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

			if err == nil {
				errUpload := c.ClientStorage.UploadPassword(application, encryptedPassword, metadataPassword, opTime, opTime, User, initVector)
				if errUpload != nil {
					c.ClientLogger.Errorln("Error while writting password data to SQLite: ", errUpload)
				}
				fmt.Printf("Your password for application %s has been successfully uploaded!\n", application)
				return
			} else if err != nil && strings.Contains(err.Error(), "no rows with application") {
				fmt.Printf("Your data is not up to date, please sync the Gophkeeper")
				return
			} else if CheckErrorType(err) {
				errUpload := c.ClientStorage.UploadPassword(application, encryptedPassword, metadataPassword, opTime, User, "", initVector)
				if errUpload != nil {
					c.ClientLogger.Errorln("Error while writting password data to SQLite: ", errUpload)
				}
				err = c.ClientStorage.SavePasswordOperation(application, User, ut.Create, nil, opTime)
				if err != nil {
					c.ClientLogger.Errorln("Error while saving info about password operation: ", err)
				}
				return
			}

			fmt.Println("Internal server error, please contact Gophkeeper administrator.")

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
				fmt.Println("Please login or register to Gophkeeper.")
				return
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

			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					c.SyncData(User)
					err = c.ClientStorage.UpdateCacheMiss(User, cacheMiss)
					if err != nil {
						c.ClientLogger.Errorln(err)
					}
				}
			}

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

				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while  updating cache miss while getting sensetive data for application %s: %s", application, err)
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
					if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
						fmt.Println("Please login to Gophkeeper!")
						return
					}
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
					uploadTime := (time.Now()).UTC()
					opTime := uploadTime.Format(time.RFC3339)
					encryptedPassword, initVector, err := c.EncryptData(passwordApp.Password)
					if err != nil {
						c.ClientLogger.Errorln("Error while encrypt sensetive data for application %s: %s", application, err)
					}
					err = c.ClientStorage.UploadPassword(application, encryptedPassword, passwordApp.MetaData, passwordApp.UploadTime, opTime, User, initVector)
					if err != nil {
						c.ClientLogger.Errorf("Error while uploading sensetive data for application %s : %s", application, err)
					}
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
				fmt.Println("Please login or register to Gophkeeper.")
				return
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

			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					c.SyncData(User)
					err = c.ClientStorage.UpdateCacheMiss(User, cacheMiss)
					if err != nil {
						c.ClientLogger.Errorln(err)
					}
				}
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

			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)

			var retryCount = 1
			_, err = clientGRPC.DeletePassword(ctx, &pb.SensetiveDataMessage{
				Identificator: application,
			})
			if err != nil {
				c.ClientLogger.Errorf("Error while deleting password in Gophkeeper for application %s: %s\n", application, err)
				if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
					fmt.Println("Please login to Gophkeeper!")
					return
				}
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

			if err != nil && CheckErrorType(err) {
				errLocal := c.ClientStorage.DeletePassword(application, User)
				if errLocal != nil {
					c.ClientLogger.Errorf("Error while deleting password for application %s from local storage: %s \n", application, err)
				}
				err = c.ClientStorage.SavePasswordOperation(application, User, ut.Delete, nil, opTime)
				if err != nil {
					c.ClientLogger.Errorln("Error while saving info about password operation: ", err)
				}
				return
			} else if err != nil && strings.Contains(err.Error(), "no rows with application") {
				fmt.Printf("Gophkeeper does not contain application %s\n", application)
				return
			} else if err != nil {
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				return
			}

			errLocal := c.ClientStorage.DeletePassword(application, User)
			if errLocal != nil {
				c.ClientLogger.Errorf("Error while deleting password for application %s from local storage: %s \n", application, err)
			}
			fmt.Printf("All sensetive data regarding to application %s was successfully removed from gophkeeper", application)
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
				fmt.Println("Please login or register to Gophkeeper.")
				return
			}

			fmt.Print("Please enter appplication, that password belongs to: ")
			application, _ = reader.ReadString('\n')
			for application == "" {
				fmt.Print("Please enter appplication, that password belongs to: ")
				application, _ = reader.ReadString('\n')
			}

			fmt.Println("Process your application sensetive data...")
			c.SendCacheData()

			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					c.SyncData(User)
					err = c.ClientStorage.UpdateCacheMiss(User, cacheMiss)
					if err != nil {
						c.ClientLogger.Errorln(err)
					}
				}
			}

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
				if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
					fmt.Println("Please login to Gophkeeper!")
					return
				}
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

			if err != nil && CheckErrorType(err) {
				uploadErr := c.ClientStorage.UploadPassword(application, encryptedPassword, passwordMetadata, opTime, User, "", initVector)
				if uploadErr != nil {
					c.ClientLogger.Errorf("Error while updating  password sensetive data for application %s: %s\n", application, err)
				}
				fields := make([]string, 0)
				if newPassword != "" {
					fields = append(fields, "password")
				}
				if passwordMetadata != "" {
					fields = append(fields, "metadata")
				}
				err = c.ClientStorage.SavePasswordOperation(application, User, ut.Update, fields, opTime)
				if err != nil {
					c.ClientLogger.Errorf("Error while saving information about update operation for sensetive data of application %s: %s\n", application, err)
				}
				return
			} else if err != nil && strings.Contains(err.Error(), "no rows with application") {
				fmt.Printf("Your data is not up to date, please sync the Gophkeeper")
				return
			} else if err != nil {
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				return
			}
			uploadErr := c.ClientStorage.UploadPassword(application, encryptedPassword, passwordMetadata, opTime, User, opTime, initVector)
			if uploadErr != nil {
				c.ClientLogger.Errorf("Error while updating  password sensetive data for application %s: %s\n", application, err)
			}

			fmt.Printf("Your password for application %s has been successfully updated!\n", application)

		},
	}

	return UpdatePassword
}

// ExecutePasswordsOperation - function for executing old operations with sensetive data for application.
func (c *Client) ExecutePasswordsOperation(operation ut.Operation, userJWT string, password *pb.PasswordMessage) error {

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
	case ut.Create:
		_, err = clientGRPC.UploadPassword(ctx, password)
		if err != nil {
			if strings.Contains(err.Error(), "no rows with application") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while inserting sensetive data for application %s: %s", password.Application, err)
				}
				return nil
			}
			return fmt.Errorf("error while uploading password: %w", err)
		}
	case ut.Update:
		_, err = clientGRPC.UpdatePassword(ctx, password)
		if err != nil {
			if strings.Contains(err.Error(), "no rows with application") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while updating sensetive data for application %s: %s", password.Application, err)
				}
				return nil
			}
			return fmt.Errorf("error while updating password: %w", err)
		}
	case ut.Delete:
		_, err = clientGRPC.DeletePassword(ctx, &pb.SensetiveDataMessage{
			Identificator: password.Application,
		})
		if err != nil {
			if strings.Contains(err.Error(), "no rows with application") {
				c.ClientLogger.Errorf("Gophkeeper does not contain application %s\n", password.Application)
				return nil
			}
			return fmt.Errorf("error while deleting password: %w", err)
		}
	}

	return nil
}
