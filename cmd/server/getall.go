package main

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Tanya1515/gophkeeper.git/cmd/proto"
)

func (s *GophkeeperServer) Sync(ctx context.Context, empt *emptypb.Empty) (*pb.DataMessage, error) {
	return nil, nil
}

func (s *GophkeeperServer) SyncFiles(empt *emptypb.Empty, fileStream grpc.ServerStreamingServer[pb.FileMessage]) error {
	return nil
}
