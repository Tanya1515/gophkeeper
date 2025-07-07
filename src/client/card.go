package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// CheckCardNumber - function, that checks if bank card number is correct.
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

// CheckDateFormat - function, that checks, if date, entered for bank card, is correct.
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

// SendBankCard creates handler for processing cli command, that sends bank card credentials.
func (c *Client) SendBankCard() *cobra.Command {
	var SendBankCard = &cobra.Command{
		Use:   "card",
		Short: "Save bank card sensetive data",
		Long:  `Save bank card sensetive data: card number, cvc code, date`,
		Run: func(cmd *cobra.Command, args []string) {
			var cardNumber, metadatabankCard, cvc, date, bankName, encryptedCvc string
			var initVector []byte

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
				fmt.Println("Please login or register to Gophkeeper.")
				return
			}

			fmt.Print("Please enter card number: ")
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
			fmt.Print("Please enter cvc code of the card: ")

			cvc, _ = reader.ReadString('\n')
			cvc = strings.TrimRight(cvc, "\n")

			encryptedCvc, initVector, err = c.EncryptData(cvc)
			if err != nil {
				c.ClientLogger.Errorln("Error while encrypt sensetive data for bank card %s: %s", cardNumber, err)
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				return
			}

			fmt.Print("Please enter card date: ")

			date, _ = reader.ReadString('\n')
			date = strings.TrimRight(date, "\n")
			for {
				if CheckDateFormat(date) {
					break
				}
				fmt.Print("Your date was invalid, please enter againin formet MM/YY: ")
				date, _ = reader.ReadString('\n')
				date = strings.TrimRight(date, "\n")
			}
			fmt.Print("Please enter bank name: ")

			bankName, _ = reader.ReadString('\n')
			bankName = strings.TrimRight(bankName, "\n")

			fmt.Print("Please enter metadata for sensetive data: ")
			metadatabankCard, _ = reader.ReadString('\n')
			metadatabankCard = strings.TrimRight(metadatabankCard, "\n")

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

			fmt.Println("Process your bank card...")

			var retryCount = 1
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)
			_, err = clientGRPC.UploadBankCard(ctx, &pb.BankCardMessage{
				CardNumber: cardNumber,
				CvcCode:    cvc,
				Data:       date,
				Bank:       bankName,
				Metadata:   metadatabankCard,
				UploadTime: opTime,
			})
			if err != nil {
				c.ClientLogger.Errorf("Error while sending bank card %s credentials: %s\n", cardNumber, err)
				if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
					fmt.Println("Please login to Gophkeeper!")
					return
				}
			}

			for err != nil && retryCount != 3 {
				_, err = clientGRPC.UploadBankCard(ctx, &pb.BankCardMessage{
					CardNumber: cardNumber,
					CvcCode:    cvc,
					Data:       date,
					Bank:       bankName,
					Metadata:   metadatabankCard,
					UploadTime: opTime,
				})
				if err != nil {
					c.ClientLogger.Errorf("Error while sending bank card %s credentials: %s\n", cardNumber, err)

				}
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}

			if err == nil {
				errUpload := c.ClientStorage.UploadBankCard(cardNumber, encryptedCvc, date, bankName, metadatabankCard, opTime, opTime, User, initVector)
				if errUpload != nil {
					c.ClientLogger.Errorf("Error while writting bank card to SQLite: %s\n", errUpload)
				}
				fmt.Println("Bank card credentials successfully have been uploaded!")
				c.SendCacheData()
				return
			} else if err != nil && strings.Contains(err.Error(), "no rows with bank card") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while inserting sensetive data for bank card %s: %s", cardNumber, err)
				}
				fmt.Printf("Your data is not up to date, please sync the Gophkeeper")
				return
			} else if CheckErrorType(err) {
				errUpload := c.ClientStorage.UploadBankCard(cardNumber, encryptedCvc, date, bankName, metadatabankCard, opTime, "", User, initVector)
				if errUpload != nil {
					c.ClientLogger.Errorf("Error while writting bank card to SQLite: %s\n", errUpload)
				}
				err = c.ClientStorage.SaveCardOperation(cardNumber, User, ut.Create, nil, opTime)
				if err != nil {
					c.ClientLogger.Errorf("Error while saving info about bank card operation: %s\n", err)
				}
				return
			}

			fmt.Println("Internal server error, please contact Gophkeeper administrator.")
		},
	}

	return SendBankCard
}

// GetBankCard creates handler for processing cli command, that gets bank card credentials.
func (c *Client) GetBankCard() *cobra.Command {
	var GetCard = &cobra.Command{
		Use:   "card",
		Short: "Get bank card credentials",
		Run: func(cmd *cobra.Command, args []string) {
			var cardNumber string

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
				fmt.Println("Please login or register to Gophkeeper.")
				return
			}

			fmt.Print("Please enter card number: ")
			cardNumber, _ = reader.ReadString('\n')
			cardNumber = strings.TrimRight(cardNumber, "\n")

			ok := CheckCardNumber(cardNumber)
			for !ok {
				fmt.Print("Card number is incorrect, please enter again: ")
				cardNumber, _ = reader.ReadString('\n')
				cardNumber = strings.TrimRight(cardNumber, "\n")
				ok = CheckCardNumber(cardNumber)
			}
			fmt.Println("Process your bank card...")
			c.SendCacheData()
			cvc, date, bankName, metadataCard, initVector, exists, err := c.ClientStorage.GetBankCard(cardNumber, User)
			if exists {
				cvcDecrypted, err := c.Crypto.DecryptData(cvc, initVector)
				if err != nil {
					c.ClientLogger.Errorf("Error while decrypting sensetive data for bank card %s: %s\n", cardNumber, err)
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
					return
				}
				fmt.Printf("Card number: %s\n", cardNumber)
				fmt.Printf("Card cvc code: %s\n", cvcDecrypted)
				fmt.Printf("Card date: %s\n", date)
				fmt.Printf("Card bank: %s\n", bankName)
				fmt.Printf("Additioanl information: %s\n", metadataCard)
			} else {
				if err != nil {
					c.ClientLogger.Errorf("Error while getting bank card %s credentials from local storage for user %s: %w", cardNumber, User, err)
				}

				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while getting sensetive data for bank card %s: %s", cardNumber, err)
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

				var retryCount = 1
				bankCard, err := clientGRPC.GetBankCardCredentials(ctx, &pb.SensetiveDataMessage{
					Identificator: cardNumber,
				})

				if err != nil {
					c.ClientLogger.Errorf("Error while getting bank card %s credentials: %s", cardNumber, err)
					if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
						fmt.Println("Please login to Gophkeeper!")
						return
					}
				}

				for err != nil && retryCount != 3 {
					bankCard, err = clientGRPC.GetBankCardCredentials(ctx, &pb.SensetiveDataMessage{
						Identificator: cardNumber,
					})
					if err != nil {
						c.ClientLogger.Errorf("Error while getting bank card %s credentials: %s", cardNumber, err)
					}
					retryCount++
					time.Sleep(time.Duration(retryCount))
				}

				if err == nil {
					fmt.Printf("Card number: %s\n", cardNumber)
					fmt.Printf("Card cvc code: %s\n", bankCard.CvcCode)
					fmt.Printf("Card date: %s\n", bankCard.Data)
					fmt.Printf("Card bank: %s\n", bankCard.Bank)
					fmt.Printf("Additioanl information: %s\n", bankCard.Metadata)
					uploadTime := (time.Now()).UTC()
					opTime := uploadTime.Format(time.RFC3339)
					encryptedCvc, initVector, err := c.EncryptData(bankCard.CvcCode)
					if err != nil {
						c.ClientLogger.Errorln("Error while encrypt sensetive data for bank card %s: %s", cardNumber, err)
					}
					err = c.ClientStorage.UploadBankCard(cardNumber, encryptedCvc, bankCard.Data, bankCard.Bank, bankCard.Metadata, bankCard.UploadTime, opTime, User, initVector)
					if err != nil {
						c.ClientLogger.Errorf("Error while uploading bank card %s credentials: %s", cardNumber, err)
					}
				} else if strings.Contains(strings.Split(err.Error(), "desc")[1], "no rows in result") {
					fmt.Printf("Sensetive data for bank card %s do not exist\n", cardNumber)
				} else {
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				}
			}

		},
	}

	return GetCard

}

// DeleteBankCard creates handler for processing cli command, that deletes bank card data.
func (c *Client) DeleteBankCard() *cobra.Command {
	var DeleteCard = &cobra.Command{
		Use:   "card",
		Short: "Delete bank card credentials",
		Run: func(cmd *cobra.Command, args []string) {
			var cardNumber string

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
				fmt.Println("Please login or register to Gophkeeper.")
				return
			}

			fmt.Print("Please enter card number, that is going to be deleted: ")
			cardNumber, _ = reader.ReadString('\n')
			cardNumber = strings.TrimRight(cardNumber, "\n")

			ok := CheckCardNumber(cardNumber)
			for !ok {
				fmt.Print("Card number is incorrect, please enter again: ")
				cardNumber, _ = reader.ReadString('\n')
				cardNumber = strings.TrimRight(cardNumber, "\n")
				ok = CheckCardNumber(cardNumber)
			}
			fmt.Println("Process your bank card...")
			c.SendCacheData()
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

			var retryCount = 1
			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)
			_, err = clientGRPC.DeleteBankCardCredentials(ctx, &pb.SensetiveDataMessage{
				Identificator: cardNumber,
			})
			if err != nil {
				c.ClientLogger.Errorf("Error while deleting bank card %s: %s", cardNumber, err)
				if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
					fmt.Println("Please login to Gophkeeper!")
					return
				}
			}

			for err != nil && retryCount != 3 {
				_, err = clientGRPC.DeleteBankCardCredentials(ctx, &pb.SensetiveDataMessage{
					Identificator: cardNumber,
				})
				retryCount++
				time.Sleep(time.Duration(retryCount))

				if err != nil {
					c.ClientLogger.Errorf("Error while deleting bank card %s: %s", cardNumber, err)
				}
			}

			if err != nil && CheckErrorType(err) {
				errLocal := c.ClientStorage.DeleteBankCard(cardNumber, User)
				if errLocal != nil {
					c.ClientLogger.Errorf("Error while deleting sensetive data for bank %s from local storage: %s \n", cardNumber, err)
				}

				err = c.ClientStorage.SaveCardOperation(cardNumber, User, ut.Delete, nil, opTime)
				if err != nil {
					c.ClientLogger.Errorln("Error while saving info about password operation: ", err)
				}
			} else if err != nil && strings.Contains(err.Error(), "no rows with bank card") {
				fmt.Printf("Gophkeeper does not contain bank card %s\n", cardNumber)
				return
			} else if err != nil {
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				return
			}

			errLocal := c.ClientStorage.DeleteBankCard(cardNumber, User)
			if errLocal != nil {
				c.ClientLogger.Errorf("Error while deleting sensetive data for bank %s from local storage: %s \n", cardNumber, err)
			}
			errLocal = c.ClientStorage.DeleteBankCard(cardNumber, User)
			if errLocal != nil {
				c.ClientLogger.Errorf("Error while deleting sensetive data for bank %s from local storage: %s \n", cardNumber, err)
			}
			fmt.Printf("All sensetive data regarding to bank card %s was successfully removed from gophkeeper\n", cardNumber)

		},
	}

	return DeleteCard

}

// UpdateBankCard creates handler for processing cli command, that updates bank card credentials.
func (c *Client) UpdateBankCard() *cobra.Command {
	var UpdateCard = &cobra.Command{
		Use:   "card",
		Short: "Update bank card credentials",
		Run: func(cmd *cobra.Command, args []string) {
			var cardNumber string
			var cvc string
			var cardDate string
			var bankName string
			var metadatabankCard, cvcEncrypted string
			var initVector []byte

			reader := bufio.NewReader(os.Stdin)

			JWTToken, err := ut.GetJWT(User)
			if err != nil && strings.Contains(err.Error(), "please login or register") {
				fmt.Print(err.Error())
				return
			} else if err != nil {
				c.ClientLogger.Errorf("Error while getting user %s credentials: %s\n", User, err)
				fmt.Println("Please login or register to Gophkeeper.")
				return
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

			fmt.Print("Please enter card date, if you would like to change it: ")
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
			fmt.Println("Process your bank card...")
			c.SendCacheData()

			if cvc != "" {
				cvcEncrypted, initVector, err = c.Crypto.EncryptData(cvc)
				if err != nil {
					c.ClientLogger.Errorf("Error while encrypting cvc code for bank card %s: %s", cardNumber, err)
					fmt.Println("Internal server error, please contact Gophkeeper administrator.")
					return
				}
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

			uploadTime := (time.Now()).UTC()
			opTime := uploadTime.Format(time.RFC3339)

			var retryCount = 1
			_, err = clientGRPC.UpdateBankCardCreds(ctx, &pb.BankCardMessage{
				CardNumber: cardNumber,
				CvcCode:    cvc,
				Data:       cardDate,
				Bank:       bankName,
				Metadata:   metadatabankCard,
				UploadTime: opTime,
			})
			if err != nil {
				c.ClientLogger.Errorf("Error while updating bank card %s credentials: %s", cardNumber, err)
				if strings.Contains(err.Error(), "error while processing JWT token: token is expired") {
					fmt.Println("Please login to Gophkeeper!")
					return
				}
			}

			for err != nil && retryCount != 3 {
				_, err = clientGRPC.UpdateBankCardCreds(ctx, &pb.BankCardMessage{
					CardNumber: cardNumber,
					CvcCode:    cvc,
					Data:       cardDate,
					Bank:       bankName,
					Metadata:   metadatabankCard,
					UploadTime: opTime,
				})

				if err != nil {
					c.ClientLogger.Errorf("Error while updating bank card %s credentials: %s", cardNumber, err)
				}
				retryCount++
				time.Sleep(time.Duration(retryCount))
			}

			if err != nil && CheckErrorType(err) {
				uploadErr := c.ClientStorage.UploadBankCard(cardNumber, cvcEncrypted, cardDate, bankName, metadatabankCard, opTime, "", User, initVector)
				if uploadErr != nil {
					c.ClientLogger.Errorf("Error while updating sensetive data for bank card %s: %s\n", cardNumber, err)
				}

				fields := make([]string, 0)
				if cvc != "" {
					fields = append(fields, "cvc")
				}
				if cardDate != "" {
					fields = append(fields, "date")
				}
				if bankName != "" {
					fields = append(fields, "bank")
				}
				if metadatabankCard != "" {
					fields = append(fields, "metadata")
				}
				err = c.ClientStorage.SaveCardOperation(cardNumber, User, ut.Update, fields, opTime)
				if err != nil {
					c.ClientLogger.Errorf("Error while saving information about update operation for sensetive data of bank card %s: %s\n", cardNumber, err)
				}
				return
			} else if err != nil && strings.Contains(err.Error(), "no rows with bank card") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while  updating cache miss while updating sensetive data for bank card %s: %s", cardNumber, err)
				}
				fmt.Printf("Your data is not up to date, please sync the Gophkeeper")
				return
			} else if err != nil {
				fmt.Println("Internal server error, please contact Gophkeeper administrator.")
				return
			}

			uploadErr := c.ClientStorage.UploadBankCard(cardNumber, cvcEncrypted, cardDate, bankName, metadatabankCard, opTime, opTime, User, initVector)
			if uploadErr != nil {
				c.ClientLogger.Errorf("Error while updating sensetive data for bank card %s: %s\n", cardNumber, err)
			}
			fmt.Printf("Your bank card %s has been successfully updated.\n", cardNumber)
		},
	}
	return UpdateCard

}

// ExecuteCardsOperations - function for executing old operations with bank card credentials.
func (c *Client) ExecuteCardsOperations(operation ut.Operation, userJWT string, bankCard *pb.BankCardMessage) error {

	md := metadata.New(map[string]string{"Authorization": userJWT})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	certPath, envExists := os.LookupEnv("CERT_PATH")
	if !(envExists) {
		certPath = "../../test_certs/"
	}

	connection, err := c.ClientConnection(certPath)
	if err != nil {
		return fmt.Errorf("error while creating GRPC connection to server: %w", err)
	}

	clientGRPC := pb.NewGophkeeperClient(connection)

	switch operation {
	case ut.Create:
		_, err = clientGRPC.UploadBankCard(ctx, bankCard)
		if err != nil {
			if strings.Contains(err.Error(), "no rows with bank card") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while inserting sensetive data for bank card %s: %s", bankCard.CardNumber, err)
				}
				return nil
			}
			return fmt.Errorf("error while uploading bank card: %w", err)
		}
	case ut.Update:
		_, err = clientGRPC.UpdateBankCardCreds(ctx, bankCard)
		if err != nil {
			if strings.Contains(err.Error(), "no rows with bank card") {
				err = c.ClientStorage.UpdateCacheMiss(User, 1)
				if err != nil {
					c.ClientLogger.Errorf("Error while updating cache miss while updating sensetive data for bank card %s: %s", bankCard.CardNumber, err)
				}
				return nil
			}
			return fmt.Errorf("error while updating bank card credentials: %w", err)
		}
	case ut.Delete:
		_, err = clientGRPC.DeleteBankCardCredentials(ctx, &pb.SensetiveDataMessage{
			Identificator: bankCard.CardNumber,
		})
		if err != nil {
			if strings.Contains(err.Error(), "no rows with bank card") {
				c.ClientLogger.Errorf("Gophkeeper does not contain bank card %s\n", bankCard.CardNumber)
				return nil
			}
			return fmt.Errorf("error while deleting bank card credentials: %w", err)
		}
	}

	return nil
}
