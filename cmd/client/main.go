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

func init() {
	rootCmd.AddCommand(client.LoginCmd)
	rootCmd.AddCommand(client.RegisterCmd)

	rootCmd.AddCommand(saveCmd)
	saveCmd.AddCommand(client.SendPassword)
	client.SendPassword.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.SendPassword.MarkFlagRequired("user")
	saveCmd.AddCommand(client.SendBankCard)
	client.SendBankCard.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.SendBankCard.MarkFlagRequired("user")
	saveCmd.AddCommand(client.SendFile)
	client.SendFile.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.SendFile.MarkFlagRequired("user")

	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(client.GetCard)
	client.GetCard.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.GetCard.MarkFlagRequired("user")
	getCmd.AddCommand(client.GetFile)
	client.GetFile.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.GetFile.MarkFlagRequired("user")
	getCmd.AddCommand(client.GetPassword)
	client.GetPassword.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.GetPassword.MarkFlagRequired("user")
	getCmd.AddCommand(client.GetUserData)
	client.GetUserData.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.GetUserData.MarkFlagRequired("user")

	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(client.DeleteFile)
	client.DeleteFile.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.DeleteFile.MarkFlagRequired("user")
	deleteCmd.AddCommand(client.DeleteCard)
	client.DeleteCard.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.DeleteCard.MarkFlagRequired("user")
	deleteCmd.AddCommand(client.DeletePassword)
	client.DeletePassword.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.DeletePassword.MarkFlagRequired("user")

	rootCmd.AddCommand(updateCmd)
	updateCmd.AddCommand(client.UpdateFile)
	client.UpdateFile.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.UpdateFile.MarkFlagRequired("user")
	updateCmd.AddCommand(client.UpdateCard)
	client.UpdateCard.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.UpdateCard.MarkFlagRequired("user")
	updateCmd.AddCommand(client.UpdatePassword)
	client.UpdatePassword.Flags().StringVarP(&client.User, "user", "u", "", "User login (required)")
	client.UpdatePassword.MarkFlagRequired("user")
}

func main() {
	err := ut.CreateJWTPath()
	if err != nil {
		fmt.Println("Error while file for JWT initialization: ", err)
	}
	Execute()
	
	sqliteCache := &sql.SQLite{}

	client := client.Client{ClientStorage: sqliteCache}

	err = client.ClientStorage.Connect()
	if err != nil {
		fmt.Println("Error while connecting to client storage: ", err)
		return 
	}
}
