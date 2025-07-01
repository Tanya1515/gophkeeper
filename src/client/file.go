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

	cs "github.com/Tanya1515/gophkeeper.git/src/client_storage"
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
			var filePath string
			var metadataFile string

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
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

			file, err := os.Open(filePath)
			if err != nil {
				c.ClientLogger.Errorf("Failed to open file: %v\n", err)
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
			if err == nil {
				defer stream.CloseAndRecv()
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
					Content:  buffer[:n],
					FileName: fileName,
					MetaData: metadataFile,
				})

				retryCount = 1
				for err != nil && retryCount != 3 {
					err = stream.Send(&pb.FileMessage{
						Content:  buffer[:n],
						FileName: fileName,
						MetaData: metadataFile,
					})
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}
			}

			if err != nil {
				err = c.ClientStorage.SaveFileOperation(fileName, User, cs.Create, nil, opTime, filePath)
				if err != nil {
					c.ClientLogger.Errorf("Error while saving info about create operation for file %s: %s\n", fileName, err)
				}
			}

			err = c.ClientStorage.UploadFile(fileName, filePath, metadataFile, opTime, User)
			if err != nil {
				c.ClientLogger.Errorf("Error while saving file %s to local client storage: %s\n", fileName, err)
			}

			fmt.Printf("File with name %s was successfully sent.", fileName)

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
			var fileName string
			var filePath string

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
			}

			for filePath == "" {
				fmt.Print("Please enter path for saving file from gophkeeper: ")
				filePath, _ = reader.ReadString('\n')
				filePath = strings.TrimRight(filePath, "\n")
			}
			fmt.Print("Please enter file to get from gophkeeper: ")
			fileName, _ = reader.ReadString('\n')
			fileName = strings.TrimRight(fileName, "\n")

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()
			metadataFile, pathToFile, exists, fileContent, err := c.ClientStorage.GetFile(fileName, User)
			if exists {
				fileToSave, err := os.Create(filePath)
				if err != nil {
					c.ClientLogger.Errorf("Error while creating file with path %s: %s\n", filePath, err)
				}
				defer fileToSave.Close()

				if len(fileContent) == 0 {
					fileExists, err := os.Open(pathToFile)
					if err != nil {
						c.ClientLogger.Errorf("Error while openning existing file %s with data: %s", pathToFile, err)
					}
					_, err = io.Copy(fileExists, fileToSave)
					if err != nil {
						c.ClientLogger.Errorf("Error while copying data from existing file %s to user file %s: %s\n", filePath, pathToFile, err)
					}
					defer fileExists.Close()
				} else {

					_, err = fileToSave.Write(fileContent)
					if err != nil {
						c.ClientLogger.Errorf("Error while writting content of file %s to path %s: %s\n", fileName, filePath, err)
					}
				}
				fmt.Printf("File %s was successfully recieved!\n", fileName)
				fmt.Println("File metadata: ", metadataFile)
			} else {
				certPath, envExists := os.LookupEnv("CERT_PATH")
				if !(envExists) {
					certPath = "../../test_certs/"
				}

				connection, err := c.ClientConnection(certPath)
				if err != nil {
					fmt.Println("Error while creating GRPC connection to server: ", err)
				}

				clientGRPC := pb.NewGophkeeperClient(connection)
				md := metadata.New(map[string]string{"Authorization": JWTToken})

				ctx := metadata.NewOutgoingContext(context.Background(), md)

				fileGetter, err := clientGRPC.GetFile(ctx, &pb.SensetiveDataMessage{
					Identificator: fileName,
				})
				if err != nil {
					c.ClientLogger.Errorf("Error while getting file %s: %s\n", fileName, err)
				}
				fileToSave, err := os.Create(filePath)
				if err != nil {
					c.ClientLogger.Errorf("Error while creating file with path %s: %s\n", filePath, err)
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

					_, err = fileToSave.Write(chunkFile.Content)
					if err != nil {
						fmt.Printf("Error while writting chunk of file %s: %s\n", fileName, err)
					}
				}

				err = fileToSave.Close()
				if err != nil {
					c.ClientLogger.Errorf("Error while closing file with path %s: %s\n", filePath, err)
				}

				c.ClientLogger.Errorf("File %s was successfully recieved!\n", fileName)
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
			var fileName string
			var filePath string
			var fileMetadata string

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
			}

			for fileName == "" {
				fmt.Print("Please enter file to update: ")
				fileName, _ = reader.ReadString('\n')
				fileName = strings.TrimRight(fileName, "\n")
			}

			for filePath == "" && fileMetadata == "" {
				fmt.Print("Please enter path for updating file in gophkeeper: ")
				filePath, _ = reader.ReadString('\n')
				filePath = strings.TrimRight(filePath, "\n")

				fmt.Print("Please enter metadata for updating file metadata in gophkeeper: ")
				fileMetadata, _ = reader.ReadString('\n')
				fileMetadata = strings.TrimRight(fileMetadata, "\n")
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()

			_, _, exists, _, _ := c.ClientStorage.GetFile(fileName, User)
			if exists {
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

				if err == nil {
					defer stream.CloseAndRecv()
				}

				if filePath != "" && err == nil {
					file, err := os.Open(filePath)
					if err != nil {
						c.ClientLogger.Errorf("Failed to open file: %s\n", err)
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
						}

						retryCount = 1
						if err = stream.Send(&pb.FileMessage{
							Content:  buffer[:n],
							FileName: fileName,
							MetaData: fileMetadata,
						}); err != nil {
							c.ClientLogger.Errorf("Error while sending file chunk: %s\n", err)
							return
						}

						for err != nil && retryCount != 3 {
							err = stream.Send(&pb.FileMessage{
								Content:  buffer[:n],
								FileName: fileName,
								MetaData: fileMetadata,
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
						Content:  []byte{},
						FileName: fileName,
						MetaData: fileMetadata,
					}); err != nil {
						c.ClientLogger.Errorf("Error while sending file chunk: %s", err)
					}
					for err != nil && retryCount != 3 {
						err = stream.Send(&pb.FileMessage{
							Content:  []byte{},
							FileName: fileName,
							MetaData: fileMetadata,
						})
						if err != nil {
							c.ClientLogger.Errorf("Error while sending file chunk: %s", err)
						}
						retryCount++
						time.Sleep(time.Duration(retryCount))
					}
				}

				if err != nil {
					fields := make([]string, 0)
					if fileMetadata != "" {
						fields = append(fields, "metadata")
					}
					if filePath != "" {
						fields = append(fields, "content")
					}

					err = c.ClientStorage.SaveFileOperation(fileName, User, cs.Update, fields, opTime, filePath)
					if err != nil {
						c.ClientLogger.Errorf("Error while saving info about create operation for file %s: %s\n", fileName, err)
					}
				}

				err = c.ClientStorage.UploadFile(fileName, filePath, fileMetadata, opTime, User)
				if err != nil {
					c.ClientLogger.Errorf("Error while uploading file %s to local storage: %s", fileName, err)
					return
				}

				fmt.Printf("File with name %s was successfully updated.", fileName)
			} else {
				fmt.Printf("File %s does not exist in Gophkeeper!", fileName)
			}

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
			}

			for fileName == "" {
				fmt.Print("Please enter file to delete: ")
				fileName, _ = reader.ReadString('\n')
				fileName = strings.TrimRight(fileName, "\n")
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()

			_, _, exists, _, _ := c.ClientStorage.GetFile(fileName, User)
			if exists {
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

				errLocal := c.ClientStorage.DeleteFile(fileName, User)
				if errLocal != nil {
					c.ClientLogger.Errorf("Error while deleting file %s from local storage: %s \n", fileName, err)
				}

				if err != nil {
					err = c.ClientStorage.SaveFileOperation(fileName, User, cs.Delete, nil, opTime, "")
					if err != nil {
						c.ClientLogger.Errorf("Error while saving delete operation about file %s: %s\n", fileName, err)
					}
					fmt.Printf("Error while removing file %s: %s\n", fileName, err)
					return
				}
				fmt.Printf("File with name %s was successfully removed from gophkeeper", fileName)

			} else {
				fmt.Printf("File %s does not exist in Gophkeeper!\n", fileName)
			}

		},
	}

	return DeleteFile
}

func (c *Client) ExecuteFilesOperations(operation cs.Operation, userJWT, uploadTime, filePath string, file *pb.FileMessage) error {

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
		if err != nil {
			return fmt.Errorf("error while closing connection to gophkeeper: %w", err)
		}
	case cs.Get:
		fileGetter, err := clientGRPC.GetFile(ctx, &pb.SensetiveDataMessage{
			Identificator: file.FileName,
		})
		if err != nil {
			return fmt.Errorf("error while getting file from gophkeeper: %w", err)
		}

		fileToSave, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("error while creating file with path %s: %w", filePath, err)
		}
		var chunkFile *pb.FileMessage
		for {
			chunkFile, err = fileGetter.Recv()
			if err != nil && err != io.EOF {
				break
			} else if err == io.EOF {
				return fmt.Errorf("error while recieving content for file %s: %w", file.FileName, err)
			}

			_, err = fileToSave.Write(chunkFile.Content)
			if err != nil {
				return fmt.Errorf("Error while writting chunk of file %s: %w", filePath, err)
			}
		}

		err = fileToSave.Close()
		if err != nil {
			return fmt.Errorf("Error while closing file with path %s: %w", file.FileName, err)
		}
	case cs.Update:

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
				fmt.Println("Error while sending file to update: ", err)
				return fmt.Errorf("error while sending file to update: %w", err)
			}
		}

		_, err = stream.CloseAndRecv()
		if err != nil && err != io.EOF {
			fmt.Println("Error while closing stream for updating file: ", err)
			return fmt.Errorf("error while closing stream for updating file: %w", err)
		}
	case cs.Delete:
		_, err = clientGRPC.DeleteFile(ctx, &pb.SensetiveDataMessage{
			Identificator: file.FileName,
		})
		if err != nil {
			fmt.Println("Error while deleting file: ", err)
			return fmt.Errorf("error while deleting file: %w", err)
		}
	}

	return nil
}
