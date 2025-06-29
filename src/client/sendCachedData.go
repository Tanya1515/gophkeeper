package client

import (
	"fmt"
	"strings"
	"sync"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// FileWorker - function for getting all info about missed operations with files and execute them.
func (c *Client) FileWorker(data <-chan UserData, resultFile chan<- OperationResult, processFile *sync.WaitGroup) {
	var operationResult OperationResult
	var wg sync.WaitGroup
	defer processFile.Done()
	for d := range data {
		fmt.Println("Start process file ", d.dataIdentificator)
		metaData, filePath, _, fileContent, err := c.ClientStorage.GetFile(d.dataIdentificator, d.user)
		if err != nil {
			fmt.Println(err)
		}
		operationsInfo, err := c.ClientStorage.GetFileOpearionsInfo(d.user, d.dataIdentificator)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(operationsInfo)
		for operationID, opInfo := range operationsInfo {
			fmt.Printf("Get  %s operation Info for file %s\n", opInfo.OperationName, d.dataIdentificator)
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
				fmt.Printf("Execute  %s operation Info for file %s\n", opInfo.OperationName, d.dataIdentificator)
				err = c.ExecuteFilesOperations(opInfo.OperationName, d.JWTtoken, opInfo.OperationTime, filePath, fileToExecute)
				if err != nil {
					operationResult = OperationResult{operationID: operationID, result: err.Error()}
				} else {
					operationResult = OperationResult{operationID: operationID, result: "Success"}
				}
				resultFile <- operationResult
			}()
		}
	}
	wg.Wait()
}

// CardWorker - function for getting all info about missed operations with cards and execute them.
func (c *Client) CardWorker(data <-chan UserData, resultCard chan<- OperationResult, processCard *sync.WaitGroup) {
	var wg sync.WaitGroup
	var operationResult OperationResult
	defer processCard.Done()
	for d := range data {
		fmt.Println("Start process bank card ", d.dataIdentificator)
		cvc, date, bankName, metadataBankCard, _, err := c.ClientStorage.GetBankCard(d.dataIdentificator, d.user)
		if err != nil {
			fmt.Println(err)
		}
		operationsInfo, err := c.ClientStorage.GetCardOperationsInfo(d.user, d.dataIdentificator)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(operationsInfo)
		for operationID, opInfo := range operationsInfo {
			fmt.Printf("Get  %s operation Info for bank card %s\n", opInfo.OperationName, d.dataIdentificator)
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
				fmt.Printf("Execute  %s operation for bank card %s\n", opInfo.OperationName, d.dataIdentificator)
				err = c.ExecuteCardsOperations(opInfo.OperationName, d.JWTtoken, opInfo.OperationTime, bankCardToExecute)
				if err != nil {
					operationResult = OperationResult{operationID: operationID, result: err.Error()}
				} else {
					operationResult = OperationResult{operationID: operationID, result: "Success"}
				}
				resultCard <- operationResult
			}()

		}
	}
	wg.Wait()
}

// PasswordWorker - function for getting all info about missed operations with passwords and execute them.
func (c *Client) PasswordWorker(data <-chan UserData, resultPassword chan<- OperationResult, processPassword *sync.WaitGroup) {
	var password, metadata string
	var err error
	var operationResult OperationResult
	var wg sync.WaitGroup
	defer processPassword.Done()
	for d := range data {
		fmt.Println("Start process application ", d.dataIdentificator)
		password, metadata, _, _, err = c.ClientStorage.GetPassword(d.dataIdentificator, d.user)
		if err != nil {
			fmt.Println(err)
		}
		operationsInfo, err := c.ClientStorage.GetPasswordOperationsInfo(d.user, d.dataIdentificator)
		if err != nil {
			fmt.Println(err)
		}
		for operationID, opInfo := range operationsInfo {
			fmt.Printf("Get  %s operation Info for application %s\n", opInfo.OperationName, d.dataIdentificator)
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
				fmt.Printf("Exxecute %s operation Info for application %s\n", opInfo.OperationName, d.dataIdentificator)
				err = c.ExecutePasswordsOperation(opInfo.OperationName, d.JWTtoken, opInfo.OperationTime, passwordToExecute)
				if err != nil {
					operationResult = OperationResult{operationID: operationID, result: err.Error()}
				} else {
					operationResult = OperationResult{operationID: operationID, result: "Success"}
				}
				resultPassword <- operationResult
			}()
		}
	}
	wg.Wait()
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
		fmt.Println("Get all card operations")
		if err != nil {
			fmt.Println(err)
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

	go func() {
		defer close(jobsFiles)
		filesOperations, err := c.ClientStorage.GetAllFileWithOperation()
		fmt.Println("Get all file operations")
		if err != nil {
			fmt.Println(err)
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

	go func() {
		defer close(jobsPasswords)
		fmt.Println("Start processing password operations")
		passwordsOperations, err := c.ClientStorage.GetAllPasswordWithOperation()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Get all password operations")
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
					fmt.Println("Error while deleting operation for card: ", err)
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
					fmt.Println("Error while deleting operation for file: ", err)
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
					fmt.Println("Error while deleting operation for password: ", err)
				}
			}
		}
	}()

	wg.Wait()
}
