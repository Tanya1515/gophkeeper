package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	client "github.com/Tanya1515/gophkeeper.git/src/client"
	sql "github.com/Tanya1515/gophkeeper.git/src/client_storage/sqlite"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your CLI application",
	Long: `A longer description that explains your CLI application in detail, 
    including available commands and their usage.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to client! Use --help for usage.")
	},
}

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save commands give an opportunity to save files, passwords and bank card credentials",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to client save interface! Use --help for usage.")
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get commands give an opportunity to get sensetive data from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to client get interface! Use --help for usage.")
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete commands give an opportunity to delete sensetive data from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to client get interface! Use --help for usage.")
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update commands give an opportunity to update sensetive data in gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to client get interface! Use --help for usage.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	err := ut.CreateJWTPath()
	if err != nil {
		fmt.Println("Error while file for JWT initialization: ", err)
	}
	Execute()

	sqliteCache := &sql.SQLite{}

	consoleClient := client.Client{ClientStorage: sqliteCache}

	err = consoleClient.ClientStorage.Connect()
	if err != nil {
		fmt.Println("Error while connecting to client storage: ", err)
		return
	}

	loginCmd := consoleClient.LoginClient()
	registerCmd := consoleClient.RegisterClient()

	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(registerCmd)

	sendPasswordCmd := consoleClient.SendPassword()
	sendBankCardCmd := consoleClient.SendBankCard()
	sendFileCmd := consoleClient.SendFile()
	rootCmd.AddCommand(saveCmd)
	saveCmd.AddCommand(sendPasswordCmd)
	sendPasswordCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	sendPasswordCmd.MarkFlagRequired("user")
	saveCmd.AddCommand(sendBankCardCmd)
	sendBankCardCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	sendBankCardCmd.MarkFlagRequired("user")
	saveCmd.AddCommand(sendFileCmd)
	sendFileCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	sendFileCmd.MarkFlagRequired("user")

	getCardCmd := consoleClient.GetBankCard()
	getFileCmd := consoleClient.GetFile()
	getPasswordCmd := consoleClient.GetPassword()
	getUserDataCmd := consoleClient.GetUserData()
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getCardCmd)
	getCardCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	getCardCmd.MarkFlagRequired("user")
	getCmd.AddCommand(getFileCmd)
	getFileCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	getFileCmd.MarkFlagRequired("user")
	getCmd.AddCommand(getPasswordCmd)
	getPasswordCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	getPasswordCmd.MarkFlagRequired("user")
	getCmd.AddCommand(getUserDataCmd)
	getUserDataCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	getUserDataCmd.MarkFlagRequired("user")

	deleteFileCmd := consoleClient.DeleteFile()
	deleteCardCmd := consoleClient.DeleteBankCard()
	deletePasswordCmd := consoleClient.DeletePassword()
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteFileCmd)
	deleteFileCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	deleteFileCmd.MarkFlagRequired("user")
	deleteCmd.AddCommand(deleteCardCmd)
	deleteCardCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	deleteCardCmd.MarkFlagRequired("user")
	deleteCmd.AddCommand(deletePasswordCmd)
	deletePasswordCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	deletePasswordCmd.MarkFlagRequired("user")

	updateFileCmd := consoleClient.UpdateFile()
	updateCardCmd := consoleClient.UpdateBankCard()
	updatePasswordCmd := consoleClient.UpdatePassword()
	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(updateFileCmd)
	updateFileCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	updateFileCmd.MarkFlagRequired("user")
	updateCmd.AddCommand(updateCardCmd)
	updateCardCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	updateCardCmd.MarkFlagRequired("user")
	updateCmd.AddCommand(updatePasswordCmd)
	updatePasswordCmd.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	updatePasswordCmd.MarkFlagRequired("user")
}
