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

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
				return
			}

			fmt.Print("Please enter metadata for sensetive data: ")
			fmt.Fscan(os.Stdin, &metadataFile)

			for filePath == "" {
				fmt.Print("Please enter absolute path for file to save: ")
				fmt.Fscan(os.Stdin, &filePath)
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()

			file, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("failed to open file: %v\n", err)
				return
			}
			defer file.Close()
			fileNameArr := strings.Split(filePath, "/")
			fileName := fileNameArr[len(fileNameArr)-1]

			certPath, envExists := os.LookupEnv("CERT_PATH")
			if !(envExists) {
				certPath = "../../test_certs/"
			}

			connection, err := ClientConnection(certPath)
			if err != nil {
				fmt.Println("Error while creating GRPC connection to server: ", err)
				return
			}

			clientGRPC := pb.NewGophkeeperClient(connection)
			md := metadata.New(map[string]string{"Authorization": JWTToken})

			ctx := metadata.NewOutgoingContext(context.Background(), md)

			stream, err := clientGRPC.UploadFile(ctx)
			if err != nil {
				fmt.Printf("error while openning GRPC stream to send file: %s\n", err)
				return
			}
			uploadTime := time.Now()
			opTime := uploadTime.Format(time.RFC3339)
			var retryCount = 1
			for err != nil && retryCount != 3 {
				stream, err = clientGRPC.UploadFile(ctx)
				if err != nil {
					fmt.Printf("error while openning GRPC stream to send file: %s", err)
					return
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
					fmt.Printf("Error while sending file chunk: %s", err)
					return
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
					fmt.Printf("Error while saving info about create operation for file %s: %s\n", fileName, err)
					return
				}
			}

			err = c.ClientStorage.UploadFile(fileName, filePath, metadataFile, opTime, User)
			if err != nil {
				fmt.Printf("Error while saving file %s to local client storage: %s\n", fileName, err)
				return
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

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
			}

			for filePath == "" {
				fmt.Print("Please enter path for saving file from gophkeeper: ")
				fmt.Fscan(os.Stdin, &filePath)
			}
			fmt.Print("Please enter file to get from gophkeeper: ")
			fmt.Fscan(os.Stdin, &fileName)

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()
			metadataFile, pathToFile, exists, fileContent, err := c.ClientStorage.GetFile(fileName, User)
			if exists {
				fileToSave, err := os.Create(filePath)
				if err != nil {
					fmt.Printf("Error while creating file with path %s: %s\n", filePath, err)
					return
				}
				defer fileToSave.Close()
				if len(fileContent) == 0 {
					fileExists, err := os.Open(pathToFile)
					if err != nil {
						fmt.Printf("Error while openning existing file %s with data: %s", pathToFile, err)
						return
					}
					_, err = io.Copy(fileExists, fileToSave)
					if err != nil {
						fmt.Printf("Error while copying data from existing file %s to user file %s: %s\n", filePath, pathToFile, err)
						return
					}
					defer fileExists.Close()
				} else {

					_, err = fileToSave.Write(fileContent)
					if err != nil {
						fmt.Printf("Error while writting content of file %s to path %s: %s\n", fileName, filePath, err)
						return
					}
				}
				fmt.Printf("File %s was successfully recieved!\n", fileName)
				fmt.Println("File metadata: ", metadataFile)
			} else {
				certPath, envExists := os.LookupEnv("CERT_PATH")
				if !(envExists) {
					certPath = "../../test_certs/"
				}

				connection, err := ClientConnection(certPath)
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
					fmt.Printf("Error while getting file %s: %s\n", fileName, err)
				}
				fileToSave, err := os.Create(filePath)
				if err != nil {
					fmt.Printf("Error while creating file with path %s: %s\n", filePath, err)
				}
				var chunkFile *pb.FileMessage
				for {
					chunkFile, err = fileGetter.Recv()
					if err != nil && err != io.EOF {
						fmt.Printf("Error while recieving new data portion of file %s: %s\n", fileName, err)
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
					fmt.Printf("Error while closing file with path %s: %s\n", filePath, err)
				}

				fmt.Printf("File %s was successfully recieved!\n", fileName)
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
				fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
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

				connection, err := ClientConnection(certPath)
				if err != nil {
					fmt.Println("Error while creating GRPC connection to server: ", err)
				}

				clientGRPC := pb.NewGophkeeperClient(connection)
				md := metadata.New(map[string]string{"Authorization": JWTToken})

				ctx := metadata.NewOutgoingContext(context.Background(), md)

				uploadTime := time.Now()
				opTime := uploadTime.Format(time.RFC3339)

				var retryCount = 1
				stream, err := clientGRPC.UpdateFile(ctx)
				if err != nil {
					fmt.Printf("error while openning GRPC stream to update file: %s\n", err)
				}

				for err != nil && retryCount != 3 {
					stream, err = clientGRPC.UpdateFile(ctx)
					if err != nil {
						fmt.Printf("error while openning GRPC stream to update file: %s\n", err)
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
						fmt.Printf("failed to open file: %v\n", err)
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
							fmt.Printf("Error while sending file chunk: %s", err)
						}

						retryCount = 1
						if err = stream.Send(&pb.FileMessage{
							Content:  buffer[:n],
							FileName: fileName,
							MetaData: fileMetadata,
						}); err != nil {
							fmt.Printf("Error while sending file chunk: %s", err)
							return
						}

						for err != nil && retryCount != 3 {
							err = stream.Send(&pb.FileMessage{
								Content:  buffer[:n],
								FileName: fileName,
								MetaData: fileMetadata,
							})
							if err != nil {
								fmt.Printf("Error while sending file chunk: %s", err)
							}
							retryCount++
							time.Sleep(time.Duration(retryCount))
						}
						if err != nil {
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
						fmt.Printf("Error while sending file chunk: %s", err)
						return
					}
					for err != nil && retryCount != 3 {
						err = stream.Send(&pb.FileMessage{
							Content:  []byte{},
							FileName: fileName,
							MetaData: fileMetadata,
						})
						retryCount++
						time.Sleep(time.Duration(retryCount))
					}
				}
				fmt.Println(err)
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
						fmt.Printf("Error while saving info about create operation for file %s: %s\n", fileName, err)
					}
				}

				err = c.ClientStorage.UploadFile(fileName, filePath, fileMetadata, opTime, User)
				if err != nil {
					fmt.Printf("Error while uploading file %s to local storage: %s", fileName, err)
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

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
			}

			for fileName == "" {
				fmt.Print("Please enter file to delete: ")
				fmt.Fscan(os.Stdin, &fileName)
			}

			fmt.Println("Start processing file with sensetive data...")
			c.SendCacheData()

			_, _, exists, _, _ := c.ClientStorage.GetFile(fileName, User)
			if exists {
				certPath, envExists := os.LookupEnv("CERT_PATH")
				if !(envExists) {
					certPath = "../../test_certs/"
				}

				connection, err := ClientConnection(certPath)
				if err != nil {
					fmt.Println("Error while creating GRPC connection to server: ", err)
				}

				uploadTime := time.Now()
				opTime := uploadTime.Format(time.RFC3339)
				clientGRPC := pb.NewGophkeeperClient(connection)
				md := metadata.New(map[string]string{"Authorization": JWTToken})

				ctx := metadata.NewOutgoingContext(context.Background(), md)

				var retryCount = 1
				_, err = clientGRPC.DeleteFile(ctx, &pb.SensetiveDataMessage{
					Identificator: fileName,
				})

				for err != nil && retryCount != 3 {
					_, err = clientGRPC.DeleteFile(ctx, &pb.SensetiveDataMessage{
						Identificator: fileName,
					})
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}

				errLocal := c.ClientStorage.DeleteFile(fileName, User)
				if errLocal != nil {
					fmt.Printf("Error while deleting file %s from local storage: %s \n", fileName, err)
				}

				if err != nil {
					err = c.ClientStorage.SaveFileOperation(fileName, User, cs.Delete, nil, opTime, "")
					if err != nil {
						fmt.Printf("Error while saving delete operation about file %s: %s\n", fileName, err)
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

	connection, err := ClientConnection(certPath)
	if err != nil {
		fmt.Println("Error while creating GRPC connection to server: ", err)
		return fmt.Errorf("error while creating GRPC connection to server: %w", err)
	}

	clientGRPC := pb.NewGophkeeperClient(connection)

	switch operation {
	case cs.Create:
		stream, err := clientGRPC.UploadFile(ctx)
		if err != nil {
			fmt.Printf("Error while openning GRPC stream to send file: %s\n", err)
			return fmt.Errorf("error while openning GRPC stream to send file: %w", err)
		}

		if filePath != "" {
			fileDesc, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("Failed to open file: %s\n", err)
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
					fmt.Printf("Error while reading file chunk: %s", err)
					return fmt.Errorf("error while reading file chunk: %w", err)
				}
				file.Content = buffer[:n]
				if err := stream.Send(file); err != nil {
					fmt.Printf("Error while sending file chunk: %s", err)
					return fmt.Errorf("error while sending file chunk: %w", err)
				}
			}
		} else {
			err = stream.Send(file)
			if err != nil {
				fmt.Printf("Error while sending file %s: %s\n", file.FileName, err)
				return fmt.Errorf("error while sending file %s: %w", file.FileName, err)
			}
		}

		_, err = stream.CloseAndRecv()
		if err != nil {
			fmt.Println("Error while closing connection to gophkeeper: ", err)
			return fmt.Errorf("error while closing connection to gophkeeper: %w", err)
		}
	case cs.Get:
		fileGetter, err := clientGRPC.GetFile(ctx, &pb.SensetiveDataMessage{
			Identificator: file.FileName,
		})
		if err != nil {
			fmt.Println("Error while getting file from gophkeeper: ", err)
			return fmt.Errorf("error while getting file from gophkeeper: %w", err)
		}

		fileToSave, err := os.Create(filePath)
		if err != nil {
			fmt.Printf("Error while creating file with path %s: %s\n", filePath, err)
			return fmt.Errorf("error while creating file with path %s: %w", filePath, err)
		}
		var chunkFile *pb.FileMessage
		for {
			chunkFile, err = fileGetter.Recv()
			if err != nil && err != io.EOF {
				fmt.Printf("Error while recieving new data portion of file %s: %s\n", filePath, err)
				break
			} else if err == io.EOF {
				fmt.Printf("Error while recieving content for file %s: %w\n", file.FileName, err)
				return fmt.Errorf("error while recieving content for file %s: %w", file.FileName, err)
			}

			_, err = fileToSave.Write(chunkFile.Content)
			if err != nil {
				fmt.Printf("Error while writting chunk of file %s: %s\n", filePath, err)
				return fmt.Errorf("Error while writting chunk of file %s: %w", filePath, err)
			}
		}

		err = fileToSave.Close()
		if err != nil {
			fmt.Printf("Error while closing file with path %s: %s\n", file.FileName, err)
			return fmt.Errorf("Error while closing file with path %s: %w", file.FileName, err)
		}
	case cs.Update:

		stream, err := clientGRPC.UpdateFile(ctx)
		if err != nil {
			fmt.Printf("Error while openning GRPC stream to update file: %s\n", err)
			return fmt.Errorf("error while openning GRPC stream to update file: %w", err)
		}

		if filePath != "" {
			fileDesc, err := os.Open(filePath)
			if err != nil {
				fmt.Printf("Failed to open file: %s\n", err)
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
					fmt.Printf("Error while sending file chunk: %s", err)
					return fmt.Errorf("error while reading file chunk: %w", err)
				}
				file.Content = buffer[:n]
				if err := stream.Send(file); err != nil {
					fmt.Printf("Error while sending file chunk: %s", err)
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
