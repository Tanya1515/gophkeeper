package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// SyncAllFiles - function for getting all files from server and saving them locally.
func (c *Client) SyncAllFiles(wg *sync.WaitGroup, JWTToken string, clientGRPC pb.GophkeeperClient, user string) {
	var fileName, fileMetadata, uploadTime string
	var wgFileSave sync.WaitGroup
	var fileToSave *os.File
	md := metadata.New(map[string]string{"Authorization": JWTToken})

	ctx := metadata.NewOutgoingContext(context.Background(), md)
	stream, err := clientGRPC.SyncFiles(ctx, &emptypb.Empty{})
	if err != nil {
		c.ClientLogger.Errorf("Error while grpc connection set up: %s\n", err)
		return
	}

	file, err := stream.Recv()
	if err != nil && err != io.EOF {
		c.ClientLogger.Errorf("Error while recieving file chunk %s from gophkeeper: %s\n", file.FileName, err)
		return
	}

	if fileName != file.FileName {
		fileName = file.FileName
		fileToSave, err = os.Create("/tmp/" + fileName)
		if err != nil {
			c.ClientLogger.Errorf("Error while creating file with path %s: %s\n", "/tmp/"+fileName, err)
			return
		}
		fileMetadata = file.MetaData
		uploadTime = file.UploadTime
	}

	_, err = fileToSave.Write(file.Content)
	if err != nil {
		c.ClientLogger.Errorf("Error while saving file %s: %s\n", fileName, err)
		return
	}

	for {
		file, err := stream.Recv()
		if err != nil && err != io.EOF {
			c.ClientLogger.Errorf("Error while recieving file chunk %s from gophkeeper: %s\n", file.FileName, err)
			fmt.Println("Some errors, while sync data with yor files. Please contact administrator or try again later.")
			return
		}
		if file.End {
			err = fileToSave.Close()
			if err != nil {
				c.ClientLogger.Errorf("Error while closing file %s\n", "/tmp/"+fileName)
				fmt.Println("Some errors, while sync data with yor files. Please contact administrator or try again later.")
				return
			}
			wgFileSave.Add(1)
			go func(fileName string) {
				defer wgFileSave.Done()
				syncTime := (time.Now()).UTC()
				opTime := syncTime.Format(time.RFC3339)

				err := c.ClientStorage.UploadFile(fileName, "/tmp/"+fileName, fileMetadata, uploadTime, opTime, User)
				if err != nil {
					c.ClientLogger.Errorln(err)
				}
			}(fileName)
			break
		}

		if fileName != file.FileName {
			err = fileToSave.Close()
			if err != nil {
				c.ClientLogger.Errorf("Error while closing file %s\n", "/tmp/"+fileName)
				fmt.Println("Some errors, while sync data with yor files. Please contact administrator or try again later.")
				return
			}
			wgFileSave.Add(1)

			go func(fileName string) {
				defer wgFileSave.Done()
				syncTime := (time.Now()).UTC()
				opTime := syncTime.Format(time.RFC3339)

				err := c.ClientStorage.UploadFile(fileName, "/tmp/"+fileName, fileMetadata, uploadTime, opTime, User)
				if err != nil {
					c.ClientLogger.Errorln(err)
				}
			}(fileName)
			fileName = file.FileName
			fileMetadata = file.MetaData
			fileToSave, err = os.Create("/tmp/" + fileName)
			if err != nil {
				c.ClientLogger.Errorf("Error while creating file with path %s: %s\n", "/tmp/"+fileName, err)
				fmt.Println("Some errors, while sync data with yor files. Please contact administrator or try again later.")
				return
			}
		}

		_, err = fileToSave.Write(file.Content)
		if err != nil {
			c.ClientLogger.Errorf("Error while saving file %s: %s\n", fileName, err)
			fmt.Println("Some errors, while sync data with yor files. Please contact administrator or try again later.")
			return
		}

	}
	wgFileSave.Wait()
	wg.Done()
}

// SyncData - function for application and bank card credentials synchronization.
func (c *Client) SyncData(userName string) error {
	var wg sync.WaitGroup
	JWTToken, err := ut.GetJWT(userName)
	if err != nil && strings.Contains(err.Error(), "please login or register") {
		c.ClientLogger.Error(err.Error())
		return err
	} else if err != nil {
		c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", userName, err)
		fmt.Println("Please login or register to Gophkeeper.")
		return err
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

	wg.Add(1)

	go c.SyncAllFiles(&wg, JWTToken, clientGRPC, userName)

	sensetiveData, err := clientGRPC.Sync(ctx, &emptypb.Empty{})
	if err != nil {
		c.ClientLogger.Errorf("Error while getting all sensetive data for User %s: %s\n", userName, err)
		return err
	}

	for _, passwordInfo := range sensetiveData.Passwords {
		encryptedPassword, initVector, err := c.EncryptData(passwordInfo.Password)
		if err != nil {
			c.ClientLogger.Errorln("Error while encrypt sensetive data for application %s: %s", passwordInfo.Application, err)
			continue
		}
		syncTime := (time.Now()).UTC()
		opTime := syncTime.Format(time.RFC3339)
		err = c.ClientStorage.UploadPassword(passwordInfo.Application, encryptedPassword, passwordInfo.MetaData, passwordInfo.UploadTime, opTime, userName, initVector)
		if err != nil {
			c.ClientLogger.Errorf("Error while uploading sensetive data for application: %s: %s", passwordInfo.Application, err)
			continue
		}

	}

	for _, bankCardCreds := range sensetiveData.BankCards {
		encryptedCvc, initVector, err := c.EncryptData(bankCardCreds.CvcCode)
		if err != nil {
			c.ClientLogger.Errorf("Error while encrypt sensetive data for bank card %s: %s\n", bankCardCreds.CardNumber, err)
			continue
		}
		syncTime := (time.Now()).UTC()
		opTime := syncTime.Format(time.RFC3339)
		err = c.ClientStorage.UploadBankCard(bankCardCreds.CardNumber, encryptedCvc, bankCardCreds.Data, bankCardCreds.Bank, bankCardCreds.Metadata, bankCardCreds.UploadTime, opTime, userName, initVector)
		if err != nil {
			c.ClientLogger.Errorf("Error while uploading credentials for bank card %s: %s\n", bankCardCreds.CardNumber, err)
			continue
		}
	}

	wg.Wait()
	return nil

}

// GetUserData creates handler for processing cli command, that gets all sensetive data for current user.
func (c *Client) GetUserData() *cobra.Command {

	var GetUserData = &cobra.Command{
		Use:   "all",
		Short: "Get description of all User sensetive data",
		Run: func(cmd *cobra.Command, args []string) {

			err := c.SyncData(User)
			if err != nil {
				fmt.Printf("All data for User %s was synchronyzed\n", User)
				return
			}
			fmt.Println("Something went wrong, please contact your administrator or try again later.")
		},
	}

	return GetUserData

}
