package main

import "github.com/spf13/cobra"

// доделать
var getUserData = &cobra.Command{
	Use:   "data",
	Short: "Get description of all user sensetive data",
	Run: func(cmd *cobra.Command, args []string) {
	},
}
