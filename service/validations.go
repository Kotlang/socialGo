package service

import (
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// All input validations should be added here.

func ValidateUserPostRequest(req *socialPb.UserPostRequest) error {
	if len(req.Post) == 0 {
		return status.Error(codes.InvalidArgument, "Post text is empty.")
	}

	return nil
}

func ValidateEventRequest(req *socialPb.CreateEventRequest) error {
	if len(req.Description) == 0 {
		return status.Error(codes.InvalidArgument, "Event text is empty.")
	}

	return nil
}
