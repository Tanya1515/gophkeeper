package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
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
	Short: "Get commands give an opportunity to get sensetive data from gopherkeeper",
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

var user string

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(registerCmd)

	rootCmd.AddCommand(saveCmd)
	saveCmd.AddCommand(sendPassword)
	sendPassword.Flags().StringVarP(&user, "user", "u", "", "User login (required)")
	sendPassword.MarkFlagRequired("user")
	saveCmd.AddCommand(sendBankCard)
	sendBankCard.Flags().StringVarP(&user, "user", "u", "", "User login (required)")
	sendBankCard.MarkFlagRequired("user")
	saveCmd.AddCommand(sendFile)
	sendFile.Flags().StringVarP(&user, "user", "u", "", "User login (required)")
	sendFile.MarkFlagRequired("user")

	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getCard)
	getCard.Flags().StringVarP(&user, "user", "u", "", "User login (required)")
	getCard.MarkFlagRequired("user")
	getCmd.AddCommand(getFile)
	getFile.Flags().StringVarP(&user, "user", "u", "", "User login (required)")
	getFile.MarkFlagRequired("user")
	getCmd.AddCommand(getPassword)
	getPassword.Flags().StringVarP(&user, "user", "u", "", "User login (required)")
	getPassword.MarkFlagRequired("user")
	getCmd.AddCommand(getUserData)
	getUserData.Flags().StringVarP(&user, "user", "u", "", "User login (required)")
	getUserData.MarkFlagRequired("user")
}

func main() {
	err := ut.CreateJWTPath()
	if err != nil {
		fmt.Println("Error while file for JWT initialization: ", err)
	}
	Execute()
}
