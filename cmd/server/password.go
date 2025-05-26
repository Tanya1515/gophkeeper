package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (s *GophkeeperServer) UploadPassword(ctx context.Context, passwordData *pb.PasswordMessage) (*emptypb.Empty, error) {
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.UploadPassword(ctxDB, passwordData.Password, passwordData.Application, passwordData.MetaData)
	if err != nil {
		s.Logger.Errorf("error while uploading password for user %s for application %s: %s", ctx.Value(ut.LoginKey), passwordData.Application, err)
		return nil, fmt.Errorf("error while uploading password for user %s for application %s: %w", ctx.Value(ut.LoginKey), passwordData.Application, err)
	}

	return nil, nil
}

func (s *GophkeeperServer) DeletePassword(ctx context.Context, passwordData *pb.SensetiveDataMessage) (*emptypb.Empty, error) {

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.DeletePassword(ctxDB, passwordData.Identificator)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *GophkeeperServer) GetPassword(ctx context.Context, passwordData *pb.SensetiveDataMessage) (passwordApp *pb.PasswordMessage, err error) {

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	*passwordApp, err = s.DataStorage.GetPassword(ctxDB, passwordData.Identificator)
	if err != nil {
		s.Logger.Errorln("Error while getting bank credentials for card %s: %s", passwordData.Identificator, err)
		return nil, err
	}

	return
}
