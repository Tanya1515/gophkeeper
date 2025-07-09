package client

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CheckErrorType - function for checking, if error is
func CheckErrorType(err error) bool {
	if s, ok := status.FromError(err); ok {
		if s.Code() == codes.Unavailable || s.Code() == codes.DeadlineExceeded || s.Code() == codes.Canceled {
			return true
		}
	}
	return false
}
