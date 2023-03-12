package service

import (
	pb "github.com/Kotlang/socialGo/generated"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// All input validations should be added here.

func ValidateUserPostRequest(req *pb.UserPostRequest) error {
	if len(req.Post) == 0 {
		return status.Error(codes.InvalidArgument, "Post text is empty.")
	}

	return nil
}
