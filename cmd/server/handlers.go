package main

import (
	"context"
	"fmt"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	
)

func (s *GophkeeperServer) LoginUser(ctx context.Context, in *pb.User) (*pb.Status, error) {
	var status pb.Status
	fmt.Println(in.Login, in.Password)
	status.Status = "Ok"
	return &status, nil
}

func (s *GophkeeperServer) RegisterUser(ctx context.Context, in *pb.User) (*pb.Status, error) {
	var status pb.Status
	fmt.Println(in.Login, in.Password, in.Email)
	status.Status = "Ok"
	
	return &status, nil
}
