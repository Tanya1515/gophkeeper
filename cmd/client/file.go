package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

var sendFile = &cobra.Command{
	Use:   "file",
	Short: "Save file",
	Long:  `Save file with sensetive data to gophkeeper!`,
	Run: func(cmd *cobra.Command, args []string) {
		var filePath string
		var metadataFile string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
			return
		}

		fmt.Print("Please enter absolute path for file to save: ")
		fmt.Fscan(os.Stdin, &filePath)
		fmt.Print("Please enter metadata for sensetive data: ")
		fmt.Fscan(os.Stdin, &metadataFile)

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("failed to open file: %v\n", err)
			return
		}
		defer file.Close()
		fileNameArr := strings.Split(filePath, "/")
		fileName := fileNameArr[len(fileNameArr)-1]
		connection, err := ClientConnection()
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
				MetaData: metadataFile,
			}); err != nil {
				fmt.Printf("Error while sending file chunk: %s", err)
				return
			}
		}
		_, err = stream.CloseAndRecv()
		if err != nil {
			fmt.Printf("Error while recoeving response from server: %s", err)
			return
		}

		fmt.Printf("File with name %s was successfully sent.", fileName)

	},
}

// доделать
var getFile = &cobra.Command{
	Use:   "file",
	Short: "Get file from Gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		var fileName string
		var filePath string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter file to get from gophkeeper: ")
		fmt.Fscan(os.Stdin, &fileName)
		fmt.Print("Please enter path for saving file from gophkeeper: ")
		fmt.Fscan(os.Stdin, &filePath)

		connection, err := ClientConnection()
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
			if err != nil && err == io.EOF {
				fmt.Printf("Error while recieving new data portion of file %s: %s\n", fileName, err)
				break
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

var updateFile = &cobra.Command{
	Use:   "file",
	Short: "Update existing file with sensetive data in gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var deleteFile = &cobra.Command{
	Use:   "file",
	Short: "Delete existing file with sensetive data in gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		var fileName string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter file to delete: ")
		fmt.Fscan(os.Stdin, &fileName)

		connection, err := ClientConnection()
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
