package server

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
	ut "github.com/Tanya1515/gophkeeper.git/src/utils"
)

// UploadBankCard - GRPC handler for saving new bank card credentials for current user.
func (s *GophkeeperServer) UploadBankCard(ctx context.Context, bankCardData *pb.BankCardMessage) (*emptypb.Empty, error) {

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cvcCode, initVector, err := s.EncryptData(bankCardData.CvcCode)
	if err != nil {
		s.Logger.Errorln()
	}

	uploadAt, err := time.Parse(time.RFC3339, bankCardData.UploadTime)
	if err != nil {
		s.Logger.Errorf("Error while parsing uploadTime to time.Time: %s", err)
		return nil, fmt.Errorf("error while parsing uploadTime to time.Time: %w", err)
	}
	err = s.DataStorage.UploadBankCard(ctxDB, bankCardData.CardNumber, cvcCode, bankCardData.Data, bankCardData.Bank, bankCardData.Metadata, uploadAt, initVector)
	if err != nil {
		s.Logger.Errorf("error while uploading bank card data for user %s for card number %s: %s", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
		return nil, fmt.Errorf("error while uploading bank card data for user %s for card number %s: %w", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
	}

	return nil, nil
}

// DeleteBankCardCredentials - GRPC handler for deleting bank card credetials for current user.
func (s *GophkeeperServer) DeleteBankCardCredentials(ctx context.Context, bankCardCredentials *pb.SensetiveDataMessage) (*emptypb.Empty, error) {

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.DeleteBankCard(ctxDB, bankCardCredentials.Identificator)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// GetBankCardCredentials - GRPC handler for getting bank card credentials for current user.
func (s *GophkeeperServer) GetBankCardCredentials(ctx context.Context, bankCardCredentials *pb.SensetiveDataMessage) (*pb.BankCardMessage, error) {
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	bankCardCreds, initVector, err := s.DataStorage.GetBankCardCredentials(ctxDB, bankCardCredentials.Identificator)
	if err != nil {
		s.Logger.Errorln("Error while getting bank credentials for card %s: %s", bankCardCredentials.Identificator, err)
		return nil, err
	}

	bankCardCreds.CvcCode, err = s.DecryptData(bankCardCreds.CvcCode, initVector)
	if err != nil {
		s.Logger.Errorf("Error while decrypting cvc code for bank card %s: %s", bankCardCreds.CardNumber, err)
		return nil, fmt.Errorf("error while decrypting cvc code for bank card %s: %w", bankCardCreds.CardNumber, err)
	}

	return bankCardCreds, nil
}

// UpdateBankCardCreds - GRPC handler for updating bank card credentials for current user.
func (s *GophkeeperServer) UpdateBankCardCreds(ctx context.Context, bankCardData *pb.BankCardMessage) (*emptypb.Empty, error) {
	var cvcCode string
	var initVector []byte
	var err error

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if bankCardData.CvcCode != "\n" {
		cvcCode, initVector, err = s.EncryptData(bankCardData.CvcCode)
		if err != nil {
			s.Logger.Errorf("Error while encrypting data for bank card %s: %s", bankCardData.CardNumber, err)
			return nil, fmt.Errorf("error while encrypting data for bank card %s: %w", bankCardData.CardNumber, err)
		}
	}

	uploadAt, err := time.Parse(time.RFC3339, bankCardData.UploadTime)
	if err != nil {
		s.Logger.Errorf("Error while parsing uploadTime to time.Time: %s", err)
		return nil, fmt.Errorf("error while parsing uploadTime to time.Time: %w", err)
	}

	err = s.DataStorage.UploadBankCard(ctxDB, bankCardData.CardNumber, cvcCode, bankCardData.Data, bankCardData.Bank, bankCardData.Metadata, uploadAt, initVector)
	if err != nil {
		s.Logger.Errorf("error while updating bank card data for user %s for card number %s: %s", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
		return nil, fmt.Errorf("error while updating bank card data for user %s for card number %s: %w", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
	}

	return nil, nil
}
