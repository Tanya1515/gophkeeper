package client

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

var SendPassword = &cobra.Command{
	Use:   "password",
	Short: "Save password",
	Long:  `Save password from third-party service`,
	Run: func(cmd *cobra.Command, args []string) {
		var password string
		var application string
		var metadataPassword string

		JWTToken, err := ut.GetJWT(User)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
			return
		}

		fmt.Print("Please enter password to save: ")
		fmt.Fscan(os.Stdin, &password)
		fmt.Print("Please enter appplication, that password belongs to: ")
		fmt.Fscan(os.Stdin, &application)

		for application == "" {
			fmt.Print("Please enter appplication, that password belongs to: ")
			fmt.Fscan(os.Stdin, &application)
		}

		fmt.Print("Please enter metadata for sensetive data: ")
		fmt.Fscan(os.Stdin, &metadataPassword)

		for password == "" {
			fmt.Print("Please enter password to save: ")
			fmt.Fscan(os.Stdin, &password)
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

		_, err = clientGRPC.UploadPassword(ctx, &pb.PasswordMessage{
			Password:    password,
			Application: application,
			MetaData:    metadataPassword,
		})

		if err != nil {
			fmt.Printf("Error while uploading password for application %s : %s\n", application, err)
			return
		}
		fmt.Printf("Your password for application %s has been successfully uploaded!\n", application)
	},
}

var GetPassword = &cobra.Command{
	Use:   "password",
	Short: "Get password of the application from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		var application string

		JWTToken, err := ut.GetJWT(User)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
		}

		fmt.Print("Please enter appplication, that password belongs to: ")
		fmt.Fscan(os.Stdin, &application)

		for application == "" {
			fmt.Print("Please enter appplication, that password belongs to: ")
			fmt.Fscan(os.Stdin, &application)
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

		passwordApp, err := clientGRPC.GetPassword(ctx, &pb.SensetiveDataMessage{
			Identificator: application,
		})

		if err != nil {
			fmt.Printf("Error while getting password for application %s: %s", application, err)
			return
		}
		fmt.Printf("Application: %s\n", application)
		fmt.Printf("Password: %s\n", passwordApp.Password)
		fmt.Printf("Additioanl information: %s\n", passwordApp.MetaData)
	},
}

var DeletePassword = &cobra.Command{
	Use:   "password",
	Short: "Delete password of the application from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		var application string

		JWTToken, err := ut.GetJWT(User)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
		}

		fmt.Print("Please enter appplication, that password belongs to: ")
		fmt.Fscan(os.Stdin, &application)

		for application == "" {
			fmt.Print("Please enter appplication, that password belongs to: ")
			fmt.Fscan(os.Stdin, &application)
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

		_, err = clientGRPC.DeletePassword(ctx, &pb.SensetiveDataMessage{
			Identificator: application,
		})

		if err != nil {
			fmt.Printf("Error while removing sensetive data regarding the application %s: %s", application, err)
			return
		}

		fmt.Printf("All sensetive data regarding to application %s was successfully removed from gophkeeper", application)
	},
}

var UpdatePassword = &cobra.Command{
	Use:   "password",
	Short: "Update password of the application from gophkeeper",
	Run: func(cmd *cobra.Command, args []string) {
		var application string
		var newPassword string
		var passwordMetadata string

		reader := bufio.NewReader(os.Stdin)

		JWTToken, err := ut.GetJWT(User)
		if err != nil && strings.Contains(err.Error(), "please login or register") {
			fmt.Print(err.Error())
			return
		} else if err != nil {
			fmt.Printf("Error while getting user %s credentials: %s\n", User, err)
		}

		fmt.Print("Please enter appplication, that password belongs to: ")
		application, _ = reader.ReadString('\n')
		for application == "" {
			fmt.Print("Please enter appplication, that password belongs to: ")
			application, _ = reader.ReadString('\n')
		}
		application = strings.TrimRight(application, "\n")

		fmt.Print("Please enter new password: ")
		newPassword, _ = reader.ReadString('\n')
		newPassword = strings.TrimRight(newPassword, "\n")

		fmt.Print("Please entee metadata: ")
		passwordMetadata, _ = reader.ReadString('\n')
		passwordMetadata = strings.TrimRight(passwordMetadata, "\n")

		for newPassword == "" && passwordMetadata == "" {
			fmt.Printf("Please enter password or metadata for application %s for updating", application)
			fmt.Println("Please enter new password: ")
			newPassword, _ = reader.ReadString('\n')

			fmt.Println("Please entee metadata: ")
			passwordMetadata, _ = reader.ReadString('\n')
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

		_, err = clientGRPC.UpdatePassword(ctx, &pb.PasswordMessage{
			Password:    newPassword,
			Application: application,
			MetaData:    passwordMetadata,
		})

		if err != nil {
			fmt.Printf("Error while updating password for application %s : %s\n", application, err)
			return
		}
		fmt.Printf("Your password for application %s has been successfully updated!\n", application)
	},
}
