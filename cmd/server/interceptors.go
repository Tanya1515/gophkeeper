package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	ut "github.com/Tanya1515/gophkeeper.git/cmd/utils"
)

type CustomServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (c *CustomServerStream) Context() context.Context {
	return c.ctx
}

// InterceptorCheckJWTtoken - function for checking if JWT token of incoming request is correct
// and user is authorized.
func (s *GophkeeperServer) InterceptorCheckJWTtoken(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {

	if (strings.Split(info.FullMethod, "/")[2] != "RegisterUser") && (strings.Split(info.FullMethod, "/")[2] != "LoginUser") && (strings.Split(info.FullMethod, "/")[2] != "VerificationApprove") {
		md, ok := metadata.FromIncomingContext(ctx)

		if ok {
			userJWT := md.Get("Authorization")
			userLogin, err := ut.ProcessJWTToken(userJWT[0])
			if err != nil {
				s.Logger.Errorln("Error while processing JWT token: %s", err)
				return "", fmt.Errorf("error while processing JWT token: %w", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			id, err := s.DataStorage.CheckUserJWT(ctx, userLogin)
			if err != nil {
				s.Logger.Errorln("Error while user identification: %w", err)
				return "", fmt.Errorf("error while user identification: %w", err)
			}

			ctxUserID := context.WithValue(ctx, ut.IDKey, id)
			reqCTX := context.WithValue(ctxUserID, ut.LoginKey, userLogin)

			return handler(reqCTX, req)
		} else {
			s.Logger.Errorln("Request must contain JWT token in Authorazation request title")
			return "", errors.New("request must contain JWT token in Authorazation request title")
		}
	}

	return handler(ctx, req)
}

// StreamInterceptorCheckJWTToken - function for checking if JWT token of incoming request is correct
// and user is authorized (version for GRPC stream).
func (s *GophkeeperServer) StreamInterceptorCheckJWTToken(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

	md, ok := metadata.FromIncomingContext(ss.Context())
	if ok {
		userJWT := md.Get("Authorization")
		userLogin, err := ut.ProcessJWTToken(userJWT[0])
		s.Logger.Infoln(userLogin)
		if err != nil {
			s.Logger.Errorln("Error while processing JWT token: %s", err)
			return fmt.Errorf("error while processing JWT token: %w", err)
		}

		ctx, cancel := context.WithTimeout(ss.Context(), 5*time.Second)
		defer cancel()

		id, err := s.DataStorage.CheckUserJWT(ctx, userLogin)
		if err != nil {
			s.Logger.Errorln("Error while user identification: %w", err)
			return fmt.Errorf("error while user identification: %w", err)
		}

		ctxUserID := context.WithValue(ss.Context(), ut.IDKey, id)
		reqCTX := context.WithValue(ctxUserID, ut.LoginKey, userLogin)
		customStream := &CustomServerStream{
			ServerStream: ss,
			ctx:          reqCTX,
		}
		return handler(srv, customStream)
	} else {
		s.Logger.Errorln("Request must contain JWT token in Authorazation request title")
		return errors.New("request must contain JWT token in Authorazation request title")
	}
}

// StreamInterceptorLogger - function for adding info about requests to grpc server to logging system.
// (version for GRPC stream)
func (s *GophkeeperServer) StreamInterceptorLogger(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

	start := time.Now()

	err = handler(srv, ss)

	duration := time.Since(start)
	if err != nil {
		s.Logger.Errorln(
			"Method", info.FullMethod,
			"Error while processing GRPC request: ", err,
			"Duration", duration,
		)
	} else {
		s.Logger.Infoln(
			"Method", strings.Split(info.FullMethod, "/")[2],
			"Duration", duration,
			"ReponseStatus", "OK",
		)
	}

	return
}

// InterceptorLogger - function for adding info about requests to grpc server to logging system.
func (s *GophkeeperServer) InterceptorLogger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	start := time.Now()

	resp, err = handler(ctx, req)

	duration := time.Since(start)

	if err != nil {
		s.Logger.Errorln(
			"Method", info.FullMethod,
			"Error while processing GRPC request: ", err,
			"Duration", duration,
		)
	} else {
		s.Logger.Infoln(
			"Method", strings.Split(info.FullMethod, "/")[2],
			"Duration", duration,
			"ReponseStatus", "OK",
		)
	}

	return
}
