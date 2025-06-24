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
				fmt.Printf("error while openning GRPC stream to send file: %s", err)
				return
			}
			var retryCount = 1
			for err != nil || retryCount == 3 {
				stream, err = clientGRPC.UploadFile(ctx)
				if err != nil {
					fmt.Printf("error while openning GRPC stream to send file: %s", err)
					return
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
					fmt.Printf("Error while sending file chunk: %s", err)
					return
				}

				err = stream.Send(&pb.FileMessage{
					Content:  buffer[:n],
					FileName: fileName,
					MetaData: metadataFile,
				})

				retryCount = 1
				for err != nil || retryCount == 3 {
					err = stream.Send(&pb.FileMessage{
						Content:  buffer[:n],
						FileName: fileName,
						MetaData: metadataFile,
					})
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}
			}
			// if err != nil - обрабатываем файл в SQLite
			_, err = stream.CloseAndRecv()
			if err != nil && err != io.EOF {
				fmt.Printf("Error while recieving response from server: %s", err)
				return
			}

			fmt.Printf("File with name %s was successfully sent.", fileName)

		},
	}

	return SendFile

}

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
				return
			}
			fileToSave, err := os.Create(filePath)
			if err != nil {
				fmt.Printf("Error while creating file with path %s: %s\n", filePath, err)
				return
			}
			var chunkFile *pb.FileMessage
			for {
				chunkFile, err = fileGetter.Recv()
				if err != nil && err != io.EOF {
					fmt.Printf("Error while recieving new data portion of file %s: %s\n", fileName, err)
					break
				} else if err == io.EOF {
					return
				}

				_, err = fileToSave.Write(chunkFile.Content)
				if err != nil {
					fmt.Printf("Error while writting chunk of file %s: %s\n", fileName, err)
					return
				}
			}

			err = fileToSave.Close()
			if err != nil {
				fmt.Printf("Error while closing file with path %s: %s\n", filePath, err)
				return
			}

			fmt.Printf("File %s was successfully recieved!\n", fileName)
		},
	}

	return GetFile

}

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

			stream, err := clientGRPC.UpdateFile(ctx)
			if err != nil {
				fmt.Printf("error while openning GRPC stream to update file: %s", err)
				return
			}

			if filePath != "" {
				file, err := os.Open(filePath)
				if err != nil {
					fmt.Printf("failed to open file: %v\n", err)
					return
				}
				defer file.Close()

				const chunkSize = 64 * 1024
				buffer := make([]byte, chunkSize)

				for {
					n, err := file.Read(buffer)
					if err == io.EOF {
						break
					} else if err != nil {
						fmt.Printf("Error while sending file chunk: %s", err)
						return
					}

					if err := stream.Send(&pb.FileMessage{
						Content:  buffer[:n],
						FileName: fileName,
						MetaData: fileMetadata,
					}); err != nil {
						fmt.Printf("Error while sending file chunk: %s", err)
						return
					}
				}
			} else {
				if err := stream.Send(&pb.FileMessage{
					Content:  []byte{},
					FileName: fileName,
					MetaData: fileMetadata,
				}); err != nil {
					fmt.Printf("Error while sending file chunk: %s", err)
					return
				}
			}

			_, err = stream.CloseAndRecv()
			if err != nil && err != io.EOF {
				fmt.Printf("Error while recieving response from server: %s", err)
				return
			}

			fmt.Printf("File with name %s was successfully updated.", fileName)

		},
	}

	return UpdateFile

}

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

			_, err = clientGRPC.DeleteFile(ctx, &pb.SensetiveDataMessage{
				Identificator: fileName,
			})

			if err != nil {
				fmt.Printf("Error while removing file %s: %s", fileName, err)
				return
			}

			fmt.Printf("File with name %s was successfully removed from gophkeeper", fileName)
		},
	}

	return DeleteFile
}
