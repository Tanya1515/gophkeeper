package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

func (s *GophkeeperServer) UploadFile(inStream grpc.ClientStreamingServer[pb.FileMessage, emptypb.Empty]) error {

	fileToSave, err := os.CreateTemp("/tmp/", "gophkeeper")
	if err != nil {
		s.Logger.Errorln("Error while creating temporary file: %s", err)
		return fmt.Errorf("error while creating temporary file: %s", err)
	}

	ctx := inStream.Context()
	chunkFile, err := inStream.Recv()
	if err != nil {
		s.Logger.Errorln("Error while recieving file from GRPC stream: ", err)
		return fmt.Errorf("error while recieving file from GRPC stream: %w", err)
	}
	fileName := chunkFile.FileName
	fileMetadata := chunkFile.MetaData
	fileToSave.Write(chunkFile.Content)

	for {
		chunkFile, err = inStream.Recv()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			s.Logger.Errorln("Error while recieving file chunk from GRPC stream: ", err)
			return fmt.Errorf("error while recieving file chunk from GRPC stream: %w", err)
		}

		fileToSave.Write(chunkFile.Content)
	}
	tempFileName := fileToSave.Name()
	err = fileToSave.Close()
	if err != nil {
		s.Logger.Errorf("Error while closing temporary file %s: %s\n", tempFileName, err)
		return fmt.Errorf("error while closing temporary file %s: %w", tempFileName, err)
	}

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = s.DataStorage.UploadFile(ctxDB, fileName, fileMetadata)
	if err != nil {
		s.Logger.Errorln(err)
		return err
	}

	ctxFileStore, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = s.FileStorage.UploadFile(ctxFileStore, fileName, tempFileName)
	if err != nil {
		s.Logger.Errorln(err)
		return err
	}

	err = os.Remove(tempFileName)
	if err != nil {
		s.Logger.Errorf("Error while removing temporary file %s: %s", tempFileName, err)
		return fmt.Errorf("error while removing temporary file %s: %w", tempFileName, err)
	}

	err = inStream.SendMsg("success")
	if err != nil {
		s.Logger.Errorln("Error while sending closing message to client: %s", err)
		return fmt.Errorf("error while sending closing message to client: %w", err)
	}
	return nil
}

func (s *GophkeeperServer) DeleteFile(ctx context.Context, fileData *pb.SensetiveDataMessage) (*emptypb.Empty, error) {

	ctxDB, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := s.DataStorage.DeleteFile(ctxDB, fileData.Identificator)
	if err != nil {
		return nil, fmt.Errorf("error while removing file %s for user %s in database storage: %w", fileData.Identificator, ctx.Value(ut.LoginKey), err)
	}

	ctxStore, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = s.FileStorage.DeleteFile(ctxStore, fileData.Identificator)
	if err != nil {
		return nil, fmt.Errorf("error while removing file %s for user %s in file storage: %w", fileData.Identificator, ctx.Value(ut.LoginKey), err)
	}

	return nil, nil
}

// *"github.com/Tanya1515/gophkeeper.git/cmd/proto".SensetiveDataMessage, grpc.ServerStreamingServer["github.com/Tanya1515/gophkeeper.git/cmd/proto".FileMessage]

func (s *GophkeeperServer) GetFile(dataMessage *pb.SensetiveDataMessage, fileStream grpc.ServerStreamingServer[pb.FileMessage]) error {
	const chunkSize = 64 * 1024

	buffer := make([]byte, chunkSize)

	ctx := fileStream.Context()
	ctxStore, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	fileName := dataMessage.Identificator

	fileMessage := pb.FileMessage{
		FileName: fileName,
	}

	fileByte, err := s.FileStorage.GetFile(ctxStore, fileName)
	if err != nil {
		s.Logger.Errorf("Error while getting file %s from Minio: %s\n", fileName, err)
		return err
	}
	amount := len(fileByte) % 1024

	for i := 0; i < amount; i++ {
		buffer = fileByte[i : i+1024]
		s.Logger.Infoln("send data: %s", string(buffer))
		fileMessage.Content = buffer
		err = fileStream.Send(&fileMessage)
		if err != nil {
			s.Logger.Errorf("Error while sending file %s chunk: %s\n", fileName, err)
			return err
		}
	}

	return nil
}
