package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(registerCmd)
}

func main() {
	err := CreateJWTPath()
	if err != nil {
		fmt.Println("Error while file for JWT initialization: ", err)
	}
	Execute()
}
