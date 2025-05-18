package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *GophkeeperServer) LoginUser(ctx context.Context, in *pb.User) (*emptypb.Empty, error) {
	fmt.Println(in.Login, in.Password)

	ctxDB, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	_, err := s.DataStorage.LoginUser(ctxDB, in.Login, in.Password)
	if err != nil {
		return nil, err
	}

	// если логин и пароль корректные
	// отправить OTP - отдельная горутина?

	return &emptypb.Empty{}, nil
}

func (s *GophkeeperServer) RegisterUser(ctx context.Context, in *pb.User) (*emptypb.Empty, error) {

	err := s.FileStorage.CreateUserFileStorage(in.Login)
	if err != nil {
		s.Logger.Errorln(err)
		return nil, err
	}

	ctxDB, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()

	err = s.DataStorage.RegisterUser(ctxDB, in.Login, in.Password, in.Email)
	if err != nil {
		return nil, err
	}

	// generate OTP
	// send OTP on email - отдельная горутина?

	return &emptypb.Empty{}, nil
}
