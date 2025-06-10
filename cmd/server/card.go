package main

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (s *GophkeeperServer) UploadBankCard(ctx context.Context, bankCardData *pb.BankCardMessage) (*emptypb.Empty, error) {
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cvcCode := s.EncryptData(bankCardData.CvcCode)

	err := s.DataStorage.UploadBankCard(ctxDB, bankCardData.CardNumber, cvcCode, bankCardData.Data, bankCardData.Bank, bankCardData.Metadata)
	if err != nil {
		s.Logger.Errorf("error while uploading bank card data for user %s for card number %s: %s", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
		return nil, fmt.Errorf("error while uploading bank card data for user %s for card number %s: %w", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
	}

	return nil, nil
}

func (s *GophkeeperServer) DeleteBankCardCredentials(ctx context.Context, bankCardCredentials *pb.SensetiveDataMessage) (*emptypb.Empty, error) {

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.DeleteBankCard(ctxDB, bankCardCredentials.Identificator)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *GophkeeperServer) GetBankCardCredentials(ctx context.Context, bankCardCredentials *pb.SensetiveDataMessage) (*pb.BankCardMessage, error) {
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	bankCardCreds, err := s.DataStorage.GetBankCardCredentials(ctxDB, bankCardCredentials.Identificator)
	if err != nil {
		s.Logger.Errorln("Error while getting bank credentials for card %s: %s", bankCardCredentials.Identificator, err)
		return nil, err
	}

	bankCardCreds.CvcCode, err = s.DecryptData(bankCardCreds.CvcCode)
	if err != nil {
		s.Logger.Errorf("Error while decrypting cvc code for bank card %s: %s", bankCardCreds.CardNumber, err)
		return nil, fmt.Errorf("error while decrypting cvc code for bank card %s: %w", bankCardCreds.CardNumber, err)
	}

	return bankCardCreds, nil
}

func (s *GophkeeperServer) UpdateBankCardCreds(ctx context.Context, bankCardData *pb.BankCardMessage) (*emptypb.Empty, error){
	var cvcCode string
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if bankCardData.CvcCode != "" {
		cvcCode = s.EncryptData(bankCardData.CvcCode)
	}

	err := s.DataStorage.UpdateBankCardCreds(ctxDB, bankCardData.CardNumber, cvcCode, bankCardData.Data, bankCardData.Bank, bankCardData.Metadata)
	if err != nil {
		s.Logger.Errorf("error while updating bank card data for user %s for card number %s: %s", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
		return nil, fmt.Errorf("error while updating bank card data for user %s for card number %s: %w", ctx.Value(ut.LoginKey), bankCardData.CardNumber, err)
	}

	return nil, nil
}
