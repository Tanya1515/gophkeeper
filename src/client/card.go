package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func CheckCardNumber(cardNumber string) bool {
	var num, sum int64
	arrayDigits := make([]int64, 0, 16)

	res64, err := strconv.ParseInt(cardNumber, 10, 64)
	if err != nil {
		panic(err)
	}

	for res64 > 0 {
		num = res64 % 10
		res64 = res64 / 10
		arrayDigits = append(arrayDigits, num)
	}

	if len(arrayDigits) < 16 {
		return false
	}

	ok := (len(arrayDigits)) % 2
	for key, value := range arrayDigits {
		if (ok == 1) && ((key % 2) == ok) && (key != 0) {
			value = value * 2
			if value > 9 {
				value = value - 9
			}
		}

		if (ok == 0) && (((key + 1) % 2) == ok) && (key != 0) {
			value = value * 2
			if value > 9 {
				value = value - 9
			}
		}
		sum += value
	}

	if sum%10 == 0 {
		return true
	} else {
		return false
	}
}

func CheckDateFormat(date string) (ok bool) {
	bankDateResult := strings.Split(date, "/")

	if bankDateResult[0] == "01" || bankDateResult[0] == "02" || bankDateResult[0] == "03" || bankDateResult[0] == "04" || bankDateResult[0] == "05" || bankDateResult[0] == "06" || bankDateResult[0] == "07" || bankDateResult[0] == "08" || bankDateResult[0] == "09" || bankDateResult[0] == "10" || bankDateResult[0] == "11" || bankDateResult[0] == "12" {
		bankDateYear, err := strconv.Atoi(bankDateResult[1])

		if err != nil {
			return false
		}
		if (bankDateYear >= 1) || (bankDateYear <= 99) {
			return true
		}
	}
	return false
}

var sendBankCard = &cobra.Command{
	Use:   "card",
	Short: "Save bank card sensetive data",
	Long:  `Save bank card sensetive data: card number, cvc code, date`,
	Run: func(cmd *cobra.Command, args []string) {
		var cardNumber string
		var cvc string
		var date string
		var bankName string
		var metadatabankCard string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter card number: ")
		fmt.Fscan(os.Stdin, &cardNumber)
		for {
			if CheckCardNumber(cardNumber) {
				break
			}
			fmt.Print("Your card number was invalid, please enter again: ")
			fmt.Fscan(os.Stdin, &cardNumber)
		}
		fmt.Print("Please enter cvc code of the card: ")
		fmt.Fscan(os.Stdin, &cvc)
		fmt.Print("Please enter card date: ")
		fmt.Fscan(os.Stdin, &date)
		for {
			if CheckDateFormat(date) {
				break
			}
			fmt.Print("Your date was invalid, please enter againin formet MM/YY: ")
			fmt.Fscan(os.Stdin, &date)
		}
		fmt.Print("Please enter bank name: ")
		fmt.Fscan(os.Stdin, &bankName)
		fmt.Print("Please enter metadata for sensetive data: ")
		fmt.Fscan(os.Stdin, &metadatabankCard)

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
		_, err = clientGRPC.UploadBankCard(ctx, &pb.BankCardMessage{
			CardNumber: cardNumber,
			CvcCode:    cvc,
			Data:       date,
			Bank:       bankName,
			Metadata:   metadatabankCard,
		})

		if err != nil {
			fmt.Printf("Error while uploading bank card credentials with card number %s\n", cardNumber)
			return
		}
		fmt.Println("Bank card credentials successfully have been uploaded!")
	},
}

var getCard = &cobra.Command{
	Use:   "card",
	Short: "Get bank card credentials",
	Run: func(cmd *cobra.Command, args []string) {
		var cardNumber string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter card number: ")
		fmt.Fscan(os.Stdin, &cardNumber)

		ok := CheckCardNumber(cardNumber)
		for !ok {
			fmt.Print("Card number is incorrect, please enter again: ")
			fmt.Fscan(os.Stdin, &cardNumber)
			ok = CheckCardNumber(cardNumber)
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

		bankCard, err := clientGRPC.GetBankCardCredentials(ctx, &pb.SensetiveDataMessage{
			Identificator: cardNumber,
		})

		if err != nil {
			fmt.Printf("Error while getting bank card %s: %s", cardNumber, err)
			return
		}
		fmt.Printf("Card number: %s\n", cardNumber)
		fmt.Printf("Card cvc code: %s\n", bankCard.CvcCode)
		fmt.Printf("Card date: %s\n", bankCard.Data)
		fmt.Printf("Card bank: %s\n", bankCard.Bank)
		fmt.Printf("Additioanl information: %s\n", bankCard.Metadata)
	},
}

var deleteCard = &cobra.Command{
	Use:   "card",
	Short: "Delete bank card credentials",
	Run: func(cmd *cobra.Command, args []string) {
		var cardNumber string

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter card number, that is going to be deleted: ")
		fmt.Fscan(os.Stdin, &cardNumber)
		ok := CheckCardNumber(cardNumber)
		for !ok {
			fmt.Print("Card number is incorrect, please enter again: ")
			fmt.Fscan(os.Stdin, &cardNumber)
			ok = CheckCardNumber(cardNumber)
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

		_, err = clientGRPC.DeleteBankCardCredentials(ctx, &pb.SensetiveDataMessage{
			Identificator: cardNumber,
		})

		if err != nil {
			fmt.Printf("Error while removing sensetive data regarding the bank card %s: %s", cardNumber, err)
			return
		}

		fmt.Printf("All sensetive data regarding to bank card %s was successfully removed from gophkeeper", cardNumber)
	},
}

var updateCard = &cobra.Command{
	Use:   "card",
	Short: "Update bank card credentials",
	Run: func(cmd *cobra.Command, args []string) {
		var cardNumber string
		var cvc string
		var cardDate string
		var bankName string
		var metadatabankCard string

		reader := bufio.NewReader(os.Stdin)

		JWTToken, err := ut.GetJWT(user)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Print("Internal error")
		}

		fmt.Print("Please enter card number you would like to change: ")
		cardNumber, _ = reader.ReadString('\n')
		cardNumber = strings.TrimRight(cardNumber, "\n")
		for {
			if CheckCardNumber(cardNumber) {
				break
			}
			fmt.Print("Your card number was invalid, please enter again: ")
			cardNumber, _ = reader.ReadString('\n')
			cardNumber = strings.TrimRight(cardNumber, "\n")
		}

		fmt.Print("Please enter cvc code of the card, if you would like to change it: ")
		cvc, _ = reader.ReadString('\n')

		fmt.Print("Please enter CARD date, if you would like to change it: ")
		cardDate, _ = reader.ReadString('\n')
		cardDate = strings.TrimRight(cardDate, "\n")
		for {
			if CheckDateFormat(cardDate) || (cardDate == "") {
				break
			}
			fmt.Print("Your date was invalid, please enter againin formet MM/YY: ")
			cardDate, _ = reader.ReadString('\n')
			cardDate = strings.TrimRight(cardDate, "\n")
		}

		fmt.Print("Please enter bank name, if you would like to change it: ")
		bankName, _ = reader.ReadString('\n')
		fmt.Print("Please enter metadata for sensetive data, if you would like to change it: ")
		metadatabankCard, _ = reader.ReadString('\n')
		for metadatabankCard == "\n" && bankName == "\n" && cardDate == "\n" && cvc == "\n" {
			fmt.Print("Please enter cvc code of the card, if you would like to change it: ")
			cvc, _ = reader.ReadString('\n')
			fmt.Print("Please enter card date, if you would like to change it: ")
			cardDate, _ = reader.ReadString('\n')
			for {
				if CheckDateFormat(cardDate) {
					break
				}
				fmt.Print("Your date was invalid, please enter againin formet MM/YY: ")
				cardDate, _ = reader.ReadString('\n')
			}
			fmt.Print("Please enter bank name, if you would like to change it: ")
			bankName, _ = reader.ReadString('\n')
			fmt.Print("Please enter metadata for sensetive data, if you would like to change it: ")
			metadatabankCard, _ = reader.ReadString('\n')
		}

		cvc = strings.TrimRight(cvc, "\n")
		bankName = strings.TrimRight(bankName, "\n")
		metadatabankCard = strings.TrimRight(metadatabankCard, "\n")

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

		_, err = clientGRPC.UpdateBankCardCreds(ctx, &pb.BankCardMessage{
			CardNumber: cardNumber,
			CvcCode:    cvc,
			Data:       cardDate,
			Bank:       bankName,
			Metadata:   metadatabankCard,
		})

		if err != nil {
			fmt.Printf("Error while updating bank card credentials %s: %s\n", cardNumber, err)
			return
		}

		fmt.Printf("Your card credentials %s have been successfully updated!", cardNumber)

	},
}
