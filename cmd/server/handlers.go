package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"math/rand/v2"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (s *GophkeeperServer) LoginUser(ctx context.Context, in *pb.User) (*emptypb.Empty, error) {

	ctxDB, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	email, err := s.DataStorage.LoginUser(ctxDB, in.Login, in.Password)
	if err != nil {
		s.Logger.Errorln("Error while authentificate user %s with error: %s", in.Login, err)
		return nil, err
	}

	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())+1))

	generatedOTP := strconv.Itoa(rnd.IntN(1000000))

	err = sendOTP(email, generatedOTP)
	if err != nil {
		// process error in another way
		s.Logger.Errorln("Error while sending OTP to user's %s email %s", in.Login, email)
		return nil, err
	}

	s.Mutex.Lock()

	s.UserOTP[in.Login] = generatedOTP

	defer s.Mutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (s *GophkeeperServer) RegisterUser(ctx context.Context, in *pb.User) (*emptypb.Empty, error) {

	err := s.FileStorage.CreateUserFileStorage(in.Login)
	if err != nil {
		s.Logger.Errorln(err)
		return nil, err
	}

	ctxDB, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = s.DataStorage.RegisterUser(ctxDB, in.Login, in.Password, in.Email)
	if err != nil {
		return nil, err
	}

	rnd := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano())+1))

	generatedOTP := strconv.Itoa(rnd.IntN(1000000))

	err = sendOTP(in.Email, generatedOTP)
	if err != nil {
		// process error in another way
		s.Logger.Errorln("Error while sending OTP to user's %s email %s", in.Login, in.Email)
		return nil, err
	}

	s.Mutex.Lock()

	s.UserOTP[in.Login] = generatedOTP

	defer s.Mutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (s *GophkeeperServer) VerificationApprove(ctx context.Context, in *pb.Verify) (*pb.Result, error) {
	var result pb.Result
	var err error

	s.Mutex.Lock()

	recievedOTP := s.UserOTP[in.Login]

	s.Mutex.Unlock()

	if recievedOTP == in.OneTimePass {
		result.JWTtoken, err = ut.GenerateToken(in.Login)
		if err != nil {
			return nil, err
		}
	} else {
		s.Logger.Errorln("Error one-time password and recieved password does not match for user %s", in.Login)
		return nil, fmt.Errorf("error one-time password and recieved password does not match for user %s", in.Login)
	}

	return &result, nil
}

func (s *GophkeeperServer) UploadPassword(ctx context.Context, passwordData *pb.UploadPasswordMessage) (*emptypb.Empty, error) {
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.UploadPassword(ctxDB, passwordData.Password, passwordData.Application, passwordData.MetaData)
	if err != nil {
		s.Logger.Errorf("error while uploading password for user %s for application %s: %s", ctx.Value(ut.LogginKey), passwordData.Application, err)
		return nil, fmt.Errorf("error while uploading password for user %s for application %s: %w", ctx.Value(ut.LogginKey), passwordData.Application, err)
	}

	return nil, nil
}

func (s *GophkeeperServer) UploadBankCard(ctx context.Context, bankCardData *pb.UploadBankCardMessage) (*emptypb.Empty, error) {
	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.UploadBankCard(ctxDB, bankCardData.CardNumber, bankCardData.CvcCode, bankCardData.Data, bankCardData.Bank, bankCardData.Metadata)
	if err != nil {
		s.Logger.Errorf("error while uploading bank card data for user %s for card number %s: %s", ctx.Value(ut.LogginKey), bankCardData.CardNumber, err)
		return nil, fmt.Errorf("error while uploading bank card data for user %s for card number %s: %w", ctx.Value(ut.LogginKey), bankCardData.CardNumber, err)
	}

	return nil, nil
}

func (s *GophkeeperServer) UploadFile(in grpc.ClientStreamingServer[pb.UploadFileMessage, emptypb.Empty]) error {
	return nil
}
