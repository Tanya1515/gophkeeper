package main

import (
	"github.com/spf13/cobra"
)

// доделать
var getUserData = &cobra.Command{
	Use:   "all",
	Short: "Get description of all user sensetive data",
	Run: func(cmd *cobra.Command, args []string) {
		// JWTToken, err := ut.GetJWT(user)
		// if err != nil && strings.Contains(err.Error(), "please login or register") {
		// 	fmt.Print(err.Error())
		// 	return
		// } else if err != nil {
		// 	fmt.Print("Internal error")
		// 	return
		// }

	},
}
