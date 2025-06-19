package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

// UploadPassword - GRPC handler for uploading user password.
func (s *GophkeeperServer) UploadPassword(ctx context.Context, passwordData *pb.PasswordMessage) (*emptypb.Empty, error) {
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	s.Logger.Infoln("Recieved password: ", passwordData.Password)
	password, initVector := s.EncryptData(passwordData.Password)
	s.Logger.Infoln("Encoded password: ", password)
	err := s.DataStorage.UploadPassword(ctxDB, password, passwordData.Application, passwordData.MetaData, initVector)
	if err != nil {
		s.Logger.Errorf("error while uploading password for user %s for application %s: %s", ctx.Value(ut.LoginKey), passwordData.Application, err)
		return nil, fmt.Errorf("error while uploading password for user %s for application %s: %w", ctx.Value(ut.LoginKey), passwordData.Application, err)
	}

	return nil, nil
}

// DeletePassword - GRPC handler for deleting user password.
func (s *GophkeeperServer) DeletePassword(ctx context.Context, passwordData *pb.SensetiveDataMessage) (*emptypb.Empty, error) {

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.DeletePassword(ctxDB, passwordData.Identificator)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetPassword - GRPC handler, that returns user password and its' metadata.
func (s *GophkeeperServer) GetPassword(ctx context.Context, passwordData *pb.SensetiveDataMessage) (*pb.PasswordMessage, error) {
	var err error
	var passwordApp pb.PasswordMessage
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	passwordApp, initVector, err := s.DataStorage.GetPassword(ctxDB, passwordData.Identificator)
	if err != nil {
		s.Logger.Errorln("Error while getting password for application %s: %s", passwordData.Identificator, err)
		return nil, err
	}
	passwordApp.Password, err = s.DecryptData(passwordApp.Password, initVector)
	if err != nil {
		s.Logger.Errorln("Error while decrypting password for application %s: %s", passwordApp.Application, err)
		return nil, fmt.Errorf("error while decrypting password for application %s: %w", passwordApp.Application, err)
	}

	return &passwordApp, err
}

// UpdatePassword - function, that updates user password or its' metadata.
func (s *GophkeeperServer) UpdatePassword(ctx context.Context, passwordData *pb.PasswordMessage) (*emptypb.Empty, error) {
	var password string
	var initVector []byte
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if passwordData.Password != "" {
		password, initVector = s.EncryptData(passwordData.Password)
	}

	err := s.DataStorage.UpdatePassword(ctxDB, password, passwordData.Application, passwordData.MetaData, initVector)
	if err != nil {
		s.Logger.Errorf("error while updating password for user %s for application %s: %s", ctx.Value(ut.LoginKey), passwordData.Application, err)
		return nil, fmt.Errorf("error while updating password for user %s for application %s: %w", ctx.Value(ut.LoginKey), passwordData.Application, err)
	}

	return nil, nil
}
