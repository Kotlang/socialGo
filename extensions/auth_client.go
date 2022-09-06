package extensions

import (
	"context"

	pb "github.com/Kotlang/socialGo/generated"
	"github.com/SaiNageswarS/go-api-boot/logger"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AuthClient struct {
	grpcContext context.Context
}

func NewAuthClient(grpcContext context.Context) *AuthClient {
	return &AuthClient{
		grpcContext: grpcContext,
	}
}

func (c *AuthClient) GetAuthorProfile(userId string) chan *pb.UserProfileProto {
	result := make(chan *pb.UserProfileProto)

	go func() {
		conn, err := grpc.Dial("20.193.225.77:50051", grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			logger.Error("Failed getting connection with auth service", zap.Error(err))
			result <- nil
		}
		defer conn.Close()

		client := pb.NewProfileClient(conn)

		jwtToken, err := grpc_auth.AuthFromMD(c.grpcContext, "bearer")
		if err != nil {
			logger.Error("Failed getting jwt token", zap.Error(err))
			result <- nil
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "Authorization", "bearer "+jwtToken)
		resp, err := client.GetProfileById(
			ctx,
			&pb.GetProfileRequest{UserId: userId})
		if err != nil {
			logger.Error("Failed getting user profile", zap.Error(err))
			result <- nil
		}

		result <- resp
	}()

	return result
}
