package client

import (
	"fmt"
	"strings"

	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

func (c *Client) CardWorker(data <-chan UserData, result chan<- string) {
	for d := range data {

	}
}

func (c *Client) FileWorker(data <-chan UserData, result chan<- string) {
	for d := range data {

	}
}

func (c *Client) PasswordWorker(data <-chan UserData, result chan<- string) {
	var password, metadata string
	for d := range data {
		password, metadata, err := c.ClientStorage.GetPassword(d.dataIdentificator, d.user)
		if err != nil {
			fmt.Println(err)
		}

		c.ExecutePasswordsOperation()
	}
}

func (c *Client) SendCacheData() {
	cardOperations, err := c.ClientStorage.GetAllCardWithOperation()
	if err != nil {
		return
	}

	jobsCards := make(chan UserData, 10)
	jobsFiles := make(chan UserData, 10)
	jobsPasswords := make(chan UserData, 10)
	resultsCards := make(chan string, 10)
	resultsFiles := make(chan string, 10)
	resultsPasswords := make(chan string, 10)

	for j := 1; j <= 10; j++ {
		go c.CardWorker(jobsCards, resultsCards)
		go c.FileWorker(jobsFiles, resultsFiles)
		go c.PasswordWorker(jobsPasswords, resultsPasswords)
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

	close(jobsPasswords)
	close(jobsCards)
	close(jobsFiles)

	return

}
