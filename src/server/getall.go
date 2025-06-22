//Server - package for storing clients' sensetive data
// such as passwords, bank cards, files. 
package server

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/src/proto"
)

// Sync - GRPC handler for getting all passwords and bank card credentials for current user.
func (s *GophkeeperServer) Sync(ctx context.Context, empt *emptypb.Empty) (*pb.DataMessage, error) {

	var wg sync.WaitGroup
	var result pb.DataMessage

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	passwordsInfo := make([]*pb.PasswordMessage, 0, 100)

	passwordVectors, err := s.DataStorage.GetAllPasswords(ctxDB, &passwordsInfo)
	if err != nil {
		s.Logger.Errorln(err)
		return nil, err
	}

	wg.Add(1)
	go func() {
		for _, passwordData := range passwordsInfo {
			passwordData.Password, err = s.DecryptData(passwordData.Password, passwordVectors[passwordData.Password])
			if err != nil {
				s.Logger.Errorln(err)
			}
		}
		defer wg.Done()
	}()

	bankCards := make([]*pb.BankCardMessage, 0, 100)

	cvcVectors, err := s.DataStorage.GetAllCardsCredentials(ctxDB, &bankCards)

	if err != nil {
		s.Logger.Errorln(err)
		return nil, err
	}

	wg.Add(1)
	go func() {
		for _, cardData := range bankCards {
			cardData.CvcCode, err = s.DecryptData(cardData.CvcCode, cvcVectors[cardData.CvcCode])
			if err != nil {
				s.Logger.Errorln(err)
			}
		}
		defer wg.Done()
	}()

	wg.Wait()
	result.Passwords = passwordsInfo
	result.BankCards = bankCards

	return &result, nil
}

// SyncFiles - GRPC handler for getting all files for current user.
func (s *GophkeeperServer) SyncFiles(empt *emptypb.Empty, fileStream grpc.ServerStreamingServer[pb.FileMessage]) error {
	const chunkSize = 64 * 1024
	ctx := fileStream.Context()

	buffer := make([]byte, chunkSize)

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	files, err := s.DataStorage.GetAllFilesInfo(ctxDB)
	if err != nil {
		s.Logger.Errorln(err)
		return err
	}

	for fileName, metadata := range files {
		ctxStore, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		fileMessage := pb.FileMessage{
			FileName: fileName,
			MetaData: metadata,
		}

		fileByte, err := s.FileStorage.GetFile(ctxStore, fileName)
		if err != nil {
			s.Logger.Errorf("Error while getting file %s from Minio: %s\n", fileName, err)
			return err
		}
		amount := len(fileByte) / 1024

		for i := 0; i <= amount; i++ {
			if (len(fileByte) < i*1024) || (amount == 0) {
				buffer = fileByte[i*1024:]
			} else {
				buffer = fileByte[i : i+1024]
			}
			fileMessage.Content = buffer
			err = fileStream.Send(&fileMessage)
			if err != nil {
				s.Logger.Errorf("Error while sending file %s chunk: %s\n", fileName, err)
				return err
			}
		}
	}

	fileMessage := pb.FileMessage{
		End: true,
	}

	err = fileStream.Send(&fileMessage)
	if err != nil {
		s.Logger.Errorf("Error while sending last message: %w \n", err)
		return err
	}

	return nil
}
