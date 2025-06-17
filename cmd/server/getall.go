package main

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

func (s *GophkeeperServer) Sync(ctx context.Context, empt *emptypb.Empty) (*pb.DataMessage, error) {
	var result pb.DataMessage

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	passwordsInfo := make([]*pb.PasswordMessage, 0, 100)

	passwordVectors, err := s.DataStorage.GetAllPasswords(ctxDB, passwordsInfo)
	if err != nil {
		s.Logger.Errorln(err)
		return nil, err
	}

	// переделать в горутину
	for _, passwordData := range passwordsInfo {
		passwordData.Password, err = s.DecryptData(passwordData.Password, passwordVectors[passwordData.Password])
		if err != nil {
			s.Logger.Errorln(err)
			return nil, err
		}
	}

	bankCards := make([]*pb.BankCardMessage, 0, 100)

	result.Passwords = passwordsInfo

	cvcVectors, err := s.DataStorage.GetAllCardsCredentials(ctxDB, bankCards)

	if err != nil {
		s.Logger.Errorln(err)
		return nil, err
	}

	// переделать в горутину
	for _, cardData := range bankCards {
		cardData.CvcCode, err = s.DecryptData(cardData.CvcCode, cvcVectors[cardData.CvcCode])
		if err != nil {
			s.Logger.Errorln(err)
			return nil, err
		}
	}

	result.BankCards = bankCards

	return &result, nil
}

func (s *GophkeeperServer) SyncFiles(empt *emptypb.Empty, fileStream grpc.ServerStreamingServer[pb.FileMessage]) error {

	ctx := fileStream.Context()

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := s.DataStorage.GetAllFilesInfo(ctxDB)
	if err != nil {
		s.Logger.Errorln(err)
		return err
	}

	return nil
}
