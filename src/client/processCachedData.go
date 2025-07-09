package client

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// SortOperationsByTime - function, that sorts operationd for data type and current user by time.
func SortOperationsByTime(operations *[]ut.OperationInfo) {
	sort.Slice(*operations, func(i, j int) bool {
		t1, err1 := time.Parse(time.RFC3339, (*operations)[i].OperationTime)
		t2, err2 := time.Parse(time.RFC3339, (*operations)[j].OperationTime)

		if err1 != nil || err2 != nil {
			return false
		}

		return t1.Before(t2)
	})
}

// FileWorker - function for getting all info about missed operations with files and execute them.
func (c *Client) FileWorker(data <-chan UserData, resultFile chan<- OperationResult, processFile *sync.WaitGroup) {
	var operationResult OperationResult
	defer processFile.Done()
	for d := range data {
		c.ClientLogger.Infoln("Start processing file ", d.dataIdentificator)
		metaData, filePath, _, fileContent, err := c.ClientStorage.GetFile(d.dataIdentificator, d.user)
		if err != nil {
			c.ClientLogger.Errorln(err)
		}
		operationsInfo, err := c.ClientStorage.GetFileOpearionsInfo(d.user, d.dataIdentificator)
		if err != nil {
			c.ClientLogger.Errorln(err)
		}
		SortOperationsByTime(&operationsInfo)
		for _, opInfo := range operationsInfo {
			fileToExecute := &pb.FileMessage{FileName: d.dataIdentificator}
			for _, value := range opInfo.Fields {
				if value == "content" {
					fileToExecute.Content = fileContent
				} else if value == "metadata" {
					fileToExecute.MetaData = metaData
				}
			}
			fileToExecute.UploadTime = opInfo.OperationTime
			c.ClientLogger.Infof("Execute %s operation with file %s\n", opInfo.OperationName, d.dataIdentificator)
			err = c.ExecuteFilesOperations(opInfo.OperationName, d.JWTtoken, filePath, fileToExecute)
			if err != nil {
				operationResult = OperationResult{operationID: opInfo.OperationID, result: err.Error()}
			} else {
				operationResult = OperationResult{operationID: opInfo.OperationID, result: "Success"}
			}
			resultFile <- operationResult
			c.ClientLogger.Infof("Successfuly completed %s operation with file %s\n", opInfo.OperationName, d.dataIdentificator)
		}
	}
}

// CardWorker - function for getting all info about missed operations with cards and execute them.
func (c *Client) CardWorker(data <-chan UserData, resultCard chan<- OperationResult, processCard *sync.WaitGroup) {
	var operationResult OperationResult
	var cvcDecrypted string

	defer processCard.Done()
	for d := range data {
		c.ClientLogger.Infoln("Start processing bank card credentials for ", d.dataIdentificator)
		cvc, date, bankName, metadataBankCard, initVector, _, err := c.ClientStorage.GetBankCard(d.dataIdentificator, d.user)
		if err != nil {
			c.ClientLogger.Errorln(err)
		}
		if cvc != "" {
			cvcDecrypted, err = c.Crypto.DecryptData(cvc, initVector)
			if err != nil {
				c.ClientLogger.Errorf("Error while decrypring sensetive data for bank card %s: %s", d.dataIdentificator, err)
			}
		}
		operationsInfo, err := c.ClientStorage.GetCardOperationsInfo(d.user, d.dataIdentificator)
		if err != nil {
			c.ClientLogger.Errorln(err)
		}
		SortOperationsByTime(&operationsInfo)
		for _, opInfo := range operationsInfo {
			bankCardToExecute := &pb.BankCardMessage{CardNumber: d.dataIdentificator}
			for _, value := range opInfo.Fields {
				if value == "cvc" {
					bankCardToExecute.CvcCode = cvcDecrypted
				} else if value == "date" {
					bankCardToExecute.Data = date
				} else if value == "bankName" {
					bankCardToExecute.Bank = bankName
				} else if value == "metadata" {
					bankCardToExecute.Metadata = metadataBankCard
				}
			}
			bankCardToExecute.UploadTime = opInfo.OperationTime
			c.ClientLogger.Infof("Execute %s operation with bank card %s", opInfo.OperationName, d.dataIdentificator)
			err = c.ExecuteCardsOperations(opInfo.OperationName, d.JWTtoken, bankCardToExecute)
			if err != nil {
				operationResult = OperationResult{operationID: opInfo.OperationID, result: err.Error()}
			} else {
				operationResult = OperationResult{operationID: opInfo.OperationID, result: "Success"}
			}
			resultCard <- operationResult
			c.ClientLogger.Infof("Successfuly completed %s operation with bank card %s\n", opInfo.OperationName, d.dataIdentificator)

		}
	}
}

// PasswordWorker - function for getting all info about missed operations with passwords and execute them.
func (c *Client) PasswordWorker(data <-chan UserData, resultPassword chan<- OperationResult, processPassword *sync.WaitGroup) {
	var password, metadata, decryptedPassword string
	var err error
	var operationResult OperationResult
	var initVector []byte

	defer processPassword.Done()
	for d := range data {
		c.ClientLogger.Infoln("Start processing application and its' sensetive data ", d.dataIdentificator)
		password, metadata, _, initVector, _, err = c.ClientStorage.GetPassword(d.dataIdentificator, d.user)
		if err != nil {
			c.ClientLogger.Errorln(err)
		}
		if password != "" {
			decryptedPassword, err = c.Crypto.DecryptData(password, initVector)
			if err != nil {
				c.ClientLogger.Errorf("Error while decrypting sensetive data for application %s: %s", d.dataIdentificator, err)
				return
			}
		}
		operationsInfo, err := c.ClientStorage.GetPasswordOperationsInfo(d.user, d.dataIdentificator)
		if err != nil {
			c.ClientLogger.Errorln(err)
		}
		SortOperationsByTime(&operationsInfo)
		for _, opInfo := range operationsInfo {
			passwordToExecute := &pb.PasswordMessage{Application: d.dataIdentificator}
			for _, value := range opInfo.Fields {
				if value == "password" {
					passwordToExecute.Password = decryptedPassword
				} else if value == "metadata" {
					passwordToExecute.MetaData = metadata
				}
			}
			passwordToExecute.UploadTime = opInfo.OperationTime
			c.ClientLogger.Infof("Execute %s operation Info for application %s\n", opInfo.OperationName, d.dataIdentificator)
			err = c.ExecutePasswordsOperation(opInfo.OperationName, d.JWTtoken, passwordToExecute)
			if err != nil {
				operationResult = OperationResult{operationID: opInfo.OperationID, result: err.Error()}
			} else {
				operationResult = OperationResult{operationID: opInfo.OperationID, result: "Success"}
			}
			resultPassword <- operationResult
			c.ClientLogger.Infof("Execute %s operation Info for application %s\n", opInfo.OperationName, d.dataIdentificator)
		}
	}
}

// SendCacheData - function that executes all missed operations with sensetive data for every user:
// files, passwords and bank card credentials.
func (c *Client) SendCacheData() {

	var wg sync.WaitGroup
	var processCard, processFile, processPassword sync.WaitGroup

	jobsCards := make(chan UserData, 10)
	jobsFiles := make(chan UserData, 10)
	jobsPasswords := make(chan UserData, 10)
	resultsCards := make(chan OperationResult, 100)
	resultsFiles := make(chan OperationResult, 100)
	resultsPasswords := make(chan OperationResult, 100)

	for j := 1; j <= 10; j++ {
		processCard.Add(1)
		processFile.Add(1)
		processPassword.Add(1)
		go c.CardWorker(jobsCards, resultsCards, &processCard)
		go c.FileWorker(jobsFiles, resultsFiles, &processFile)
		go c.PasswordWorker(jobsPasswords, resultsPasswords, &processPassword)
	}

	go func() {
		defer close(jobsCards)
		cardOperations, err := c.ClientStorage.GetAllCardWithOperation()
		if err != nil {
			c.ClientLogger.Errorln(err)
			return
		}
		for user, cardsNumber := range cardOperations {
			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				c.ClientLogger.Errorln(err)
			} else if err != nil {
				fmt.Printf("Error while getting user %s credentials for executing again requests with bank cards credenbtials: %s\n", User, err)
			}
			data := UserData{user, JWTToken, ""}
			for _, cardNumber := range cardsNumber {
				data.dataIdentificator = cardNumber
				jobsCards <- data
			}

		}

	}()

	go func() {
		defer close(jobsFiles)
		filesOperations, err := c.ClientStorage.GetAllFileWithOperation()
		if err != nil {
			c.ClientLogger.Errorln(err)
			return
		}

		for user, filesName := range filesOperations {
			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				c.ClientLogger.Errorln(err.Error())
			} else if err != nil {
				fmt.Printf("Error while getting user %s credentials for executing again requests with files: %s\n", User, err)
			}
			data := UserData{user, JWTToken, ""}
			for _, fileName := range filesName {
				data.dataIdentificator = fileName
				jobsFiles <- data
			}
		}
	}()

	go func() {
		defer close(jobsPasswords)
		passwordsOperations, err := c.ClientStorage.GetAllPasswordWithOperation()
		if err != nil {
			c.ClientLogger.Errorln(err)
			return
		}
		for user, passwords := range passwordsOperations {
			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				c.ClientLogger.Errorln(err)
			} else if err != nil {
				fmt.Printf("Error while getting user %s credentials for executing again requests with passwords: %s\n", User, err)
			}
			data := UserData{user, JWTToken, ""}
			for _, password := range passwords {
				data.dataIdentificator = password
				jobsPasswords <- data
			}
		}
	}()

	go func() {
		processPassword.Wait()
		close(resultsPasswords)
	}()

	go func() {
		processCard.Wait()
		close(resultsCards)
	}()

	go func() {
		processFile.Wait()
		close(resultsFiles)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for resCard := range resultsCards {
			if resCard.result == "Success" {
				err := c.ClientStorage.DeleteOperaionByID(resCard.operationID, "card")
				if err != nil {
					c.ClientLogger.Errorln("Error while deleting operation for card: ", err)
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for resFile := range resultsFiles {
			if resFile.result == "Success" {
				err := c.ClientStorage.DeleteOperaionByID(resFile.operationID, "file")
				if err != nil {
					c.ClientLogger.Errorln("Error while deleting operation for file: ", err)
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for resPassword := range resultsPasswords {
			if resPassword.result == "Success" {
				err := c.ClientStorage.DeleteOperaionByID(resPassword.operationID, "password")
				if err != nil {
					c.ClientLogger.Errorln("Error while deleting operation for password: ", err)
				}
			}
		}
	}()

	wg.Wait()
}

// ClearData - function, that clear all outdated data for the user.
func (c *Client) ClearData(userName string) {
	var wgSyncClear sync.WaitGroup

	wgSyncClear.Add(1)
	go func() {
		defer wgSyncClear.Done()
		errClear := c.ClientStorage.ClearPasswordsByDate(User)
		if errClear != nil {
			c.ClientLogger.Errorln(errClear)
		}
	}()

	wgSyncClear.Add(1)
	go func() {
		defer wgSyncClear.Done()
		filesPath, errClear := c.ClientStorage.ClearFilesByDate(User)
		if errClear != nil {
			c.ClientLogger.Errorln(errClear)
			return
		}
		for _, filePath := range filesPath {
			errClear = os.Remove(filePath)
			if errClear != nil {
				c.ClientLogger.Errorln("Error while removing file with path %s: %s", filePath, errClear)
			}
		}
	}()

	wgSyncClear.Add(1)
	go func() {
		defer wgSyncClear.Done()
		errClear := c.ClientStorage.ClearCardsByDate(User)
		if errClear != nil {
			c.ClientLogger.Errorln(errClear)
		}
	}()

	wgSyncClear.Wait()

}
