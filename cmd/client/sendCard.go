package main

import (
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
	bankDate := strings.Split(date, "/")
	if bankDate[0] == "01" || bankDate[0] == "02" || bankDate[0] == "03" || bankDate[0] == "04" || bankDate[0] == "05" || bankDate[0] == "06" || bankDate[0] == "07" || bankDate[0] == "08" || bankDate[0] == "09" || bankDate[0] == "10" || bankDate[0] == "11" || bankDate[0] == "12" {
		bankDateYear, err := strconv.Atoi(bankDate[1])
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
	Use:   "send bank card",
	Short: "Save bank card sensetive data",
	Long:  `Save bank card sensetive data: card number, cvc code, date`,
	Run: func(cmd *cobra.Command, args []string) {
		var cardNumber string
		var cvc string
		var date string
		var bankName string
		var metadatabankCard string

		JWTToken, err := ut.GetJWT(user)
		if strings.Contains(err.Error(), "please login or register") {
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

		connection, err := ClientConnection()
		if err != nil {
			fmt.Println("Error while creating GRPC connection to server: ", err)
		}

		clientGRPC := pb.NewGophkeeperClient(connection)
		md := metadata.New(map[string]string{"Authorization": JWTToken})

		ctx := metadata.NewOutgoingContext(context.Background(), md)
		_, err = clientGRPC.UploadBankCard(ctx, &pb.UploadBankCardMessage{
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
