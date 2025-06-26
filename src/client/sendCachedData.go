package client

import (
	"fmt"
	"strings"
	"sync"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

func (c *Client) CardWorker(data <-chan UserData, result chan<- OperationResult) {
	var operationResult OperationResult
	var wg sync.WaitGroup
	for d := range data {
		metaData, filePath, fileContent, err := c.ClientStorage.GetFile(d.dataIdentificator, d.user)
		if err != nil {
			fmt.Println(err)
		}
		operationsInfo, err := c.ClientStorage.GetFileOpearionsInfo(d.user, d.dataIdentificator)
		if err != nil {
			fmt.Println(err)
		}
		for operationID, opInfo := range operationsInfo {
			fileToExecute := &pb.FileMessage{FileName: d.dataIdentificator}
			for _, value := range opInfo.Fields {
				if value == "content" {
					fileToExecute.Content = fileContent
				} else if value == "metadata" {
					fileToExecute.MetaData = metaData
				}
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = c.ExecuteFilesOperations(opInfo.OperationName, d.JWTtoken, opInfo.OperationTime, filePath, fileToExecute)
				if err != nil {
					operationResult = OperationResult{operationID: operationID, result: err.Error()}
				} else {
					operationResult = OperationResult{operationID: operationID, result: "Success"}
				}
				result <- operationResult
			}()
		}
	}
	wg.Wait()
	defer close(result)
}

func (c *Client) FileWorker(data <-chan UserData, result chan<- OperationResult) {
	var wg sync.WaitGroup
	var operationResult OperationResult
	for d := range data {
		cvc, date, bankName, metadataBankCard, err := c.ClientStorage.GetBankCard(d.dataIdentificator, d.user)
		if err != nil {
			fmt.Println(err)
		}
		operationsInfo, err := c.ClientStorage.GetCardOperationsInfo(d.user, d.dataIdentificator)
		if err != nil {
			fmt.Println(err)
		}

		for operationID, opInfo := range operationsInfo {
			bankCardToExecute := &pb.BankCardMessage{CardNumber: d.dataIdentificator}
			for _, value := range opInfo.Fields {
				if value == "cvc" {
					bankCardToExecute.CvcCode = cvc
				} else if value == "date" {
					bankCardToExecute.Data = date
				} else if value == "bankName" {
					bankCardToExecute.Bank = bankName
				} else if value == "metadata" {
					bankCardToExecute.Metadata = metadataBankCard
				}
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = c.ExecuteCardsOperations(opInfo.OperationName, d.JWTtoken, opInfo.OperationTime, bankCardToExecute)
				if err != nil {
					operationResult = OperationResult{operationID: operationID, result: err.Error()}
				} else {
					operationResult = OperationResult{operationID: operationID, result: "Success"}
				}
				result <- operationResult
			}()

		}
	}
	wg.Wait()
	defer close(result)
}

func (c *Client) PasswordWorker(data <-chan UserData, result chan<- OperationResult) {
	var password, metadata string
	var err error
	var operationResult OperationResult
	var wg sync.WaitGroup

	for d := range data {
		password, metadata, err = c.ClientStorage.GetPassword(d.dataIdentificator, d.user)
		if err != nil {
			fmt.Println(err)
		}
		operationsInfo, err := c.ClientStorage.GetPasswordOperationsInfo(d.user, d.dataIdentificator)
		if err != nil {
			fmt.Println(err)
		}
		for operationID, opInfo := range operationsInfo {
			passwordToExecute := &pb.PasswordMessage{Application: d.dataIdentificator}
			for _, value := range opInfo.Fields {
				if value == "password" {
					passwordToExecute.Password = password
				} else if value == "metadata" {
					passwordToExecute.MetaData = metadata
				}
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				err = c.ExecutePasswordsOperation(opInfo.OperationName, d.JWTtoken, opInfo.OperationTime, passwordToExecute)
				if err != nil {
					operationResult = OperationResult{operationID: operationID, result: err.Error()}
				} else {
					operationResult = OperationResult{operationID: operationID, result: "Success"}
				}
				result <- operationResult
			}()
		}
	}
	wg.Wait()
	defer close(result)
}

func (c *Client) SendCacheData() {

	var wg sync.WaitGroup

	jobsCards := make(chan UserData, 10)
	jobsFiles := make(chan UserData, 10)
	jobsPasswords := make(chan UserData, 10)
	resultsCards := make(chan OperationResult, 100)
	resultsFiles := make(chan OperationResult, 100)
	resultsPasswords := make(chan OperationResult, 100)

	for j := 1; j <= 10; j++ {
		go c.CardWorker(jobsCards, resultsCards)
		go c.FileWorker(jobsFiles, resultsFiles)
		go c.PasswordWorker(jobsPasswords, resultsPasswords)
	}

	wg.Add(1)
	go func() {
		cardOperations, err := c.ClientStorage.GetAllCardWithOperation()
		if err != nil {
			return
		}
		for user, cardsNumber := range cardOperations {
			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err)
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

	wg.Add(1)
	go func() {
		filesOperations, err := c.ClientStorage.GetAllFileWithOperation()
		if err != nil {
			return
		}

		for user, filesName := range filesOperations {
			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
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

	wg.Add(1)
	go func() {
		passwordsOperations, err := c.ClientStorage.GetAllPasswordWithOperation()
		if err != nil {
			return
		}

		for user, passwords := range passwordsOperations {
			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err)
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

	close(jobsPasswords)
	close(jobsCards)
	close(jobsFiles)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for resCard := range resultsCards {
			if resCard.result == "Success" {
				c.ClientStorage.DeleteOperaionByID(resCard.operationID, "card")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for resFile := range resultsFiles {
			if resFile.result == "Success" {
				c.ClientStorage.DeleteOperaionByID(resFile.operationID, "file")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for resPassword := range resultsPasswords {
			if resPassword.result == "Success" {
				c.ClientStorage.DeleteOperaionByID(resPassword.operationID, "file")
			}
		}
	}()

	wg.Wait()
}
