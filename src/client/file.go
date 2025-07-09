package client

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// SendFile creates handler for processing cli command, that sends user file and data.
func (c *Client) SendFile() *cobra.Command {
	var SendFile = &cobra.Command{
		Use:   "file",
		Short: "Save file",
		Long:  `Save file with sensetive data to gophkeeper!`,
		Run: func(cmd *cobra.Command, args []string) {
			var filePath, metadataFile string

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

			fmt.Print("Please enter metadata for sensetive data: ")
			metadataFile, _ = reader.ReadString('\n')
			metadataFile = strings.TrimRight(metadataFile, "\n")

			for filePath == "" {
				fmt.Print("Please enter absolute path for file to save: ")
				filePath, _ = reader.ReadString('\n')
				filePath = strings.TrimRight(filePath, "\n")
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()
			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					err = c.SyncData(User)
					if err != nil {
						err = c.ClientStorage.UpdateCacheMiss(User, 0)
						if err != nil {
							c.ClientLogger.Errorln(err)
						}
					}
				}
			}

			file, err := os.Open(filePath)
			if err != nil {
				c.ClientLogger.Errorf("Failed to open file: %v\n", err)
				fmt.Println("Internal server error, please contact Gophkeeper administrator, and please check ypur file path: ", filePath)
				return
			}
			defer file.Close()
			fileNameArr := strings.Split(filePath, "/")
			fileName := fileNameArr[len(fileNameArr)-1]

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

			stream, err := clientGRPC.UploadFile(ctx)
			if err != nil {
				c.ClientLogger.Errorf("error while openning GRPC stream to send file: %s\n", err)
			}
			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)
			var retryCount = 1
			for err != nil && retryCount != 3 {
				stream, err = clientGRPC.UploadFile(ctx)
				if err != nil {
					c.ClientLogger.Errorf("Error while openning GRPC stream to send file: %s", err)
				}
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}

			const chunkSize = 64 * 1024
			buffer := make([]byte, chunkSize)

			for {
				if err != nil {
					break
				}
				n, err := file.Read(buffer)
				if err == io.EOF {
					break
				} else if err != nil {
					c.ClientLogger.Errorf("Error while reading file %s chunk: %s", filePath, err)
				}

				err = stream.Send(&pb.FileMessage{
					Content:    buffer[:n],
					FileName:   fileName,
					MetaData:   metadataFile,
					UploadTime: opTime,
				})

				retryCount = 1
				for err != nil && retryCount != 3 {
					err = stream.Send(&pb.FileMessage{
						Content:    buffer[:n],
						FileName:   fileName,
						MetaData:   metadataFile,
						UploadTime: opTime,
					})
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}
			}
			if err == nil {
				_, err = stream.CloseAndRecv()
			}

			if err == nil || err == io.EOF {
				saveErr := c.ClientStorage.UploadFile(fileName, filePath, metadataFile, opTime, opTime, User)
				if saveErr != nil {
					c.ClientLogger.Errorf("Error while saving file %s to local client storage: %s\n", fileName, err)
				}
				fmt.Printf("File with name %s was successfully sent.", fileName)
				return
			} else if strings.Contains(err.Error(), "error while processing JWT token: ") {
				fmt.Println("Please login to Gophkeeper!")
				return
			} else if strings.Contains(err.Error(), "no rows with file") {
				fmt.Printf("Your data is not up to date, please sync the Gophkeeper")
				return
			} else if CheckErrorType(err) {
				err = c.ClientStorage.SaveFileOperation(fileName, User, ut.Create, nil, opTime, filePath)
				if err != nil {
					c.ClientLogger.Errorf("Error while saving info about create operation for file %s: %s\n", fileName, err)
				}
				saveErr := c.ClientStorage.UploadFile(fileName, filePath, metadataFile, opTime, "", User)
				if saveErr != nil {
					c.ClientLogger.Errorf("Error while saving file %s to local client storage: %s\n", fileName, err)
				}
				return
			}

			fmt.Println("Internal server error, please contact Gophkeeper administrator.")
		},
	}

	return SendFile

}

// GetFile creates handler for processing cli command, that gets user file and data.
func (c *Client) GetFile() *cobra.Command {
	var GetFile = &cobra.Command{
		Use:   "file",
		Short: "Get file from Gophkeeper",
		Run: func(cmd *cobra.Command, args []string) {
			var fileName, filePath, uploadTime string

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

			for filePath == "" {
				fmt.Print("Please enter path for saving file from gophkeeper: ")
				filePath, _ = reader.ReadString('\n')
				filePath = strings.TrimRight(filePath, "\n")
			}

			for fileName == "" {
				fmt.Print("Please enter file to get from gophkeeper: ")
				fileName, _ = reader.ReadString('\n')
				fileName = strings.TrimRight(fileName, "\n")
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()

			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					err = c.SyncData(User)
					if err != nil {
						err = c.ClientStorage.UpdateCacheMiss(User, 0)
						if err != nil {
							c.ClientLogger.Errorln(err)
						}
					}
				}
			}

			metadataFile, pathToFile, exists, fileContent, err := c.ClientStorage.GetFile(fileName, User)
			if exists {
				fileToSave, err := os.Create(filePath)
				if err != nil {
					c.ClientLogger.Errorf("Error while creating file with path %s: %s\n", filePath, err)
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
					return
				}
				defer fileToSave.Close()

				if len(fileContent) == 0 {

					fileExists, err := os.Open(pathToFile)
					if err != nil {
						c.ClientLogger.Errorf("Error while openning existing file %s with data: %s", pathToFile, err)
						fmt.Println("Internal server error, please contact Gophkeeper administrator.")
						return
					}
					_, err = io.Copy(fileToSave, fileExists)
					if err != nil {
						fmt.Println(err)
						c.ClientLogger.Errorf("Error while copying data from existing file %s to user file %s: %s\n", filePath, pathToFile, err)
						fmt.Println("Internal server error, please contact Gophkeeper administrator.")
						return
					}
					defer fileExists.Close()
				} else {
					_, err = fileToSave.Write(fileContent)
					if err != nil {
						c.ClientLogger.Errorf("Error while writting content of file %s to path %s: %s\n", fileName, filePath, err)
						fmt.Println("Internal server error, please contact Gophkeeper administrator.")
						return
					}
				}
				fmt.Printf("File %s was successfully recieved!\n", fileName)
				fmt.Println("File metadata: ", metadataFile)
			} else {
				if err != nil {
					c.ClientLogger.Errorf("Error while getting file %s from local storage for user %s: %w", fileName, User, err)
				}
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while getting file %s: %s", fileName, err)
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
				retryCount := 1
				fileGetter, err := clientGRPC.GetFile(ctx, &pb.SensetiveDataMessage{
					Identificator: fileName,
				})
				if err != nil {
					c.ClientLogger.Errorf("Error while getting file %s: %s\n", fileName, err)
				}
				for err != nil && retryCount != 3 {
					fileGetter, err = clientGRPC.GetFile(ctx, &pb.SensetiveDataMessage{
						Identificator: fileName,
					})
					if err != nil {
						c.ClientLogger.Errorf("Error while getting file %s: %s\n", fileName, err)
					}
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}
				fileToSave, err := os.Create(filePath)
				if err != nil {
					c.ClientLogger.Errorf("Error while creating file with path %s: %s\n", filePath, err)
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
					return
				}
				var chunkFile *pb.FileMessage
				for {
					chunkFile, err = fileGetter.Recv()
					if err != nil && err != io.EOF {
						c.ClientLogger.Errorf("Error while recieving new data portion of file %s: %s\n", fileName, err)
						break
					} else if err == io.EOF {
						break
					}
					if metadataFile == "" {
						metadataFile = chunkFile.MetaData
					}
					if uploadTime == "" {
						uploadTime = chunkFile.UploadTime
					}

					_, err = fileToSave.Write(chunkFile.Content)
					if err != nil {
						fmt.Printf("Error while writting chunk of file %s: %s\n", fileName, err)
						fmt.Println("Internal server error, please contact Gophkeeper administrator.")
						return
					}
				}

				defer fileToSave.Close()

				if strings.Contains(err.Error(), "error while processing JWT token: ") {
					fmt.Println("Please login to Gophkeeper!")
					return
				} else if err != nil && err != io.EOF {
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
					return
				}
				syncTime := (time.Now()).UTC()
				opTime := syncTime.Format(time.RFC3339)
				err = c.ClientStorage.UploadFile(fileName, filePath, metadataFile, uploadTime, opTime, User)
				if err != nil {
					c.ClientLogger.Errorf("Error while uploading file %s to local user storage: %s", fileName, err)
				}
				fmt.Printf("File %s was successfully recieved!\n", fileName)
				fmt.Println("File metadata is: ", metadataFile)
				c.ClientLogger.Infof("File %s was successfully recieved!\n", fileName)
			}

		},
	}

	return GetFile

}

// UpdateFile creates handler for processing cli command, that updates user file and data.
func (c *Client) UpdateFile() *cobra.Command {
	var UpdateFile = &cobra.Command{
		Use:   "file",
		Short: "Update existing file with sensetive data in gophkeeper",
		Run: func(cmd *cobra.Command, args []string) {
			var fileName, fileMetadata, filePath string

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

			for fileName == "" {
				fmt.Print("Please enter file name to update: ")
				fileName, _ = reader.ReadString('\n')
				fileName = strings.TrimRight(fileName, "\n")
			}

			for filePath == "" && fileMetadata == "" {
				fmt.Print("Please enter path where new content is: ")
				filePath, _ = reader.ReadString('\n')
				filePath = strings.TrimRight(filePath, "\n")

				fmt.Print("Please enter metadata for updating file metadata in gophkeeper: ")
				fileMetadata, _ = reader.ReadString('\n')
				fileMetadata = strings.TrimRight(fileMetadata, "\n")
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()

			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					err = c.SyncData(User)
					if err != nil {
						err = c.ClientStorage.UpdateCacheMiss(User, 0)
						if err != nil {
							c.ClientLogger.Errorln(err)
						}
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
			stream, err := clientGRPC.UpdateFile(ctx)
			if err != nil {
				c.ClientLogger.Errorf("error while openning GRPC stream to update file: %s\n", err)
			}

			for err != nil && retryCount != 3 {
				stream, err = clientGRPC.UpdateFile(ctx)
				if err != nil {
					c.ClientLogger.Errorf("Error while openning GRPC stream to update file: %s\n", err)
				}
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}

			if filePath != "" && err == nil {
				file, err := os.Open(filePath)
				if err != nil {
					c.ClientLogger.Errorf("Failed to open file: %s\n", err)
					fmt.Printf("Please check if file path %s is correct: %s", filePath, err)
					return
				}

				const chunkSize = 64 * 1024
				buffer := make([]byte, chunkSize)

				for {
					if err != nil {
						break
					}
					n, errFile := file.Read(buffer)
					if errFile == io.EOF {
						break
					} else if errFile != nil {
						c.ClientLogger.Errorf("Error while reading file chunk: %s\n", err)
						fmt.Printf("Something is wrong with your file %s: %s", filePath, err)
						return
					}

					retryCount = 1
					if err = stream.Send(&pb.FileMessage{
						Content:    buffer[:n],
						FileName:   fileName,
						MetaData:   fileMetadata,
						UploadTime: opTime,
					}); err != nil {
						c.ClientLogger.Errorf("Error while sending file chunk: %s\n", err)
					}

					for err != nil && retryCount != 3 {
						err = stream.Send(&pb.FileMessage{
							Content:    buffer[:n],
							FileName:   fileName,
							MetaData:   fileMetadata,
							UploadTime: opTime,
						})
						if err != nil {
							c.ClientLogger.Errorf("Error while sending file chunk: %s", err)
						}
						retryCount++
						time.Sleep(time.Duration(retryCount))
					}
					if err != nil {
						c.ClientLogger.Errorf("Error while sending file %s to Gophkeeper: %s\n", filePath, err)
						break
					}
				}
				file.Close()
			} else if err == nil {
				retryCount = 1
				if err = stream.Send(&pb.FileMessage{
					Content:    []byte{},
					FileName:   fileName,
					MetaData:   fileMetadata,
					UploadTime: opTime,
				}); err != nil {
					c.ClientLogger.Errorf("Error while sending file chunk: %s", err)
				}
				for err != nil && retryCount != 3 {
					err = stream.Send(&pb.FileMessage{
						Content:    []byte{},
						FileName:   fileName,
						MetaData:   fileMetadata,
						UploadTime: opTime,
					})
					if err != nil {
						c.ClientLogger.Errorf("Error while sending file chunk: %s", err)
					}
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}
			}
			if err == nil {
				_, err = stream.CloseAndRecv()
			}

			if CheckErrorType(err) {
				fields := make([]string, 0)
				if fileMetadata != "" {
					fields = append(fields, "metadata")
				}
				if filePath != "" {
					fields = append(fields, "content")
				}

				err = c.ClientStorage.SaveFileOperation(fileName, User, ut.Update, fields, opTime, filePath)
				if err != nil {
					c.ClientLogger.Errorf("Error while saving info about create operation for file %s: %s\n", fileName, err)
				}

				uploadErr := c.ClientStorage.UploadFile(fileName, filePath, fileMetadata, opTime, "", User)
				if uploadErr != nil {
					c.ClientLogger.Errorf("Error while uploading file %s to local storage: %s", fileName, err)
					return
				}
				return
			} else if strings.Contains(err.Error(), "no rows with file") {
				fmt.Printf("Your data is not up to date, please sync the Gophkeeper")
				return
			} else if err != nil && strings.Contains(err.Error(), "error while processing JWT token: ") {
				fmt.Println("Please login to Gophkeeper!")
				return
			} else if err != nil && err != io.EOF {
				c.ClientLogger.Errorln("Error while updating file in Gophkeeper: ", err)
				fmt.Println("Internal Gophkeeper server error, please contact your administrator!")
				return
			}

			uploadErr := c.ClientStorage.UploadFile(fileName, filePath, fileMetadata, opTime, opTime, User)
			if uploadErr != nil && err != io.EOF {
				c.ClientLogger.Errorf("Error while uploading file %s to local storage: %s", fileName, err)
				return
			}

			fmt.Printf("File with name %s was successfully updated.", fileName)

		},
	}

	return UpdateFile

}

// DeleteFile creates handler for processing cli command, that deletes user file and data.
func (c *Client) DeleteFile() *cobra.Command {
	var DeleteFile = &cobra.Command{
		Use:   "file",
		Short: "Delete existing file with sensetive data in gophkeeper",
		Run: func(cmd *cobra.Command, args []string) {
			var fileName string

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

			for fileName == "" {
				fmt.Print("Please enter file to delete: ")
				fileName, _ = reader.ReadString('\n')
				fileName = strings.TrimRight(fileName, "\n")
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()

			c.ClearData(User)

			cacheMiss, err := c.ClientStorage.GetCacheMiss(User)
			if err != nil {
				c.ClientLogger.Errorln(err)
			} else {
				if cacheMiss >= 20 {
					err = c.SyncData(User)
					if err != nil {
						err = c.ClientStorage.UpdateCacheMiss(User, 0)
						if err != nil {
							c.ClientLogger.Errorln(err)
						}
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

			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)
			clientGRPC := pb.NewGophkeeperClient(connection)
			md := metadata.New(map[string]string{"Authorization": JWTToken})

			ctx := metadata.NewOutgoingContext(context.Background(), md)

			var retryCount = 1
			_, err = clientGRPC.DeleteFile(ctx, &pb.SensetiveDataMessage{
				Identificator: fileName,
			})
			if err != nil {
				c.ClientLogger.Error("Error while deleting file %s from Gophkeeper: %s\n", fileName, err)
				if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
					fmt.Println("Please login to Gophkeeper!")
					return
				}
			}

			for err != nil && retryCount != 3 {
				_, err = clientGRPC.DeleteFile(ctx, &pb.SensetiveDataMessage{
					Identificator: fileName,
				})
				if err != nil {
					c.ClientLogger.Error("Error while deleting file %s from Gophkeeper: %s\n", fileName, err)
				}
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}

			if err != nil && CheckErrorType(err) {
				filePath, errLocal := c.ClientStorage.DeleteFile(fileName, User)
				if errLocal != nil {
					c.ClientLogger.Errorf("Error while deleting file %s from local storage: %s \n", fileName, err)
				}

				errLocal = os.Remove(filePath)
				if errLocal != nil {
					c.ClientLogger.Errorf("Error while removing file with path %s: %s", filePath, err)
				}

				err = c.ClientStorage.SaveFileOperation(fileName, User, ut.Delete, nil, opTime, "")
				if err != nil {
					c.ClientLogger.Errorf("Error while saving delete operation about file %s: %s\n", fileName, err)
				}
				return
			} else if err != nil && strings.Contains(err.Error(), "no rows with file name") {
				fmt.Printf("Gophkeeper does not contain file %s\n", fileName)
				return
			} else if err != nil {
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				return
			}
			filePath, errLocal := c.ClientStorage.DeleteFile(fileName, User)
			if errLocal != nil {
				c.ClientLogger.Errorf("Error while deleting file %s from local storage: %s \n", fileName, errLocal)
			}

			if filePath != "" {

			}
			if filePath != "" {
				errLocal = os.Remove(filePath)
				if errLocal != nil {
					c.ClientLogger.Errorf("Error while removing file with path %s: %s", filePath, errLocal)
				}
			}

			fmt.Printf("File with name %s was successfully removed from gophkeeper", fileName)
		},
	}

	return DeleteFile
}

// ExecuteFilesOperations - function for executing old operations with user files.
func (c *Client) ExecuteFilesOperations(operation ut.Operation, userJWT, filePath string, file *pb.FileMessage) error {

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
		stream, err := clientGRPC.UploadFile(ctx)
		if err != nil {
			return fmt.Errorf("error while openning GRPC stream to send file: %w", err)
		}

		if filePath != "" {
			fileDesc, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %s", err)
			}
			defer fileDesc.Close()

			const chunkSize = 64 * 1024
			buffer := make([]byte, chunkSize)

			for {
				n, err := fileDesc.Read(buffer)
				if err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("error while reading file chunk: %w", err)
				}
				file.Content = buffer[:n]
				if err := stream.Send(file); err != nil {
					return fmt.Errorf("error while sending file chunk: %w", err)
				}
			}
		} else {
			err = stream.Send(file)
			if err != nil {
				return fmt.Errorf("error while sending file %s: %w", file.FileName, err)
			}
		}
		_, err = stream.CloseAndRecv()
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "no rows with file") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while creating file %s: %s", file.FileName, err)
				}
				return nil
			}
			return fmt.Errorf("error while closing connection to gophkeeper: %w", err)
		}
	case ut.Update:

		stream, err := clientGRPC.UpdateFile(ctx)
		if err != nil {
			return fmt.Errorf("error while openning GRPC stream to update file: %w", err)
		}

		if filePath != "" {
			fileDesc, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}
			defer fileDesc.Close()

			const chunkSize = 64 * 1024
			buffer := make([]byte, chunkSize)

			for {
				n, err := fileDesc.Read(buffer)
				if err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("error while reading file chunk: %w", err)
				}
				file.Content = buffer[:n]
				if err := stream.Send(file); err != nil {
					return fmt.Errorf("error while sending file chunk: %w", err)
				}
			}
		} else {
			err = stream.Send(file)
			if err != nil {
				return fmt.Errorf("error while sending file to update: %w", err)
			}
		}

		_, err = stream.CloseAndRecv()
		if err != nil && err != io.EOF {
			if strings.Contains(err.Error(), "no rows with file") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while updating file %s: %s", file.FileName, err)
				}
				return nil
			}
			return fmt.Errorf("error while closing stream for updating file: %w", err)
		}
	case ut.Delete:
		_, err = clientGRPC.DeleteFile(ctx, &pb.SensetiveDataMessage{
			Identificator: file.FileName,
		})
		if err != nil {
			if strings.Contains(err.Error(), "no rows with file name") {
				c.ClientLogger.Errorf("No such file %s on Gophkeeper server side", file.FileName)
				return nil
			}
			return fmt.Errorf("error while deleting file: %w", err)
		}
	}

	return nil
}
