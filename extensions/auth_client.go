package extensions

import (
	"context"
	"os"
	"sync"

	authPb "github.com/Kotlang/socialGo/generated/auth"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/SaiNageswarS/go-api-boot/logger"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var auth_client *AuthClient = &AuthClient{}

type AuthClient struct {
	cached_conn        *grpc.ClientConn
	conn_creation_lock sync.Mutex
}

func (c *AuthClient) getConnection() *grpc.ClientConn {
	c.conn_creation_lock.Lock()
	defer c.conn_creation_lock.Unlock()

	if c.cached_conn == nil || c.cached_conn.GetState().String() != "READY" {
		if val, ok := os.LookupEnv("AUTH_TARGET"); ok {
			conn, err := grpc.Dial(val, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				logger.Error("Failed getting connection with auth service", zap.Error(err))
				return nil
			}
			c.cached_conn = conn
		} else {
			logger.Error("Failed to get AUTH_TARGET env variable")
		}

	}

	return c.cached_conn
}

func GetSocialProfiles(grpcContext context.Context, userIds []string) chan []*socialPb.SocialProfile {
	result := make(chan []*socialPb.SocialProfile)

	go func() {
		if len(userIds) == 0 {
			result <- nil
			return
		}

		// call auth service.
		conn := auth_client.getConnection()
		// logger.Info("auth connection state", zap.String("state", conn.GetState().String()))

		if conn == nil {
			result <- nil
			return
		}

		client := authPb.NewProfileClient(conn)

		ctx := prepareCallContext(grpcContext)
		if ctx == nil {
			result <- nil
			return
		}

		userIdList := &authPb.BulkGetProfileRequest{
			UserIds: userIds,
		}

		resp, err := client.BulkGetProfileByIds(ctx, userIdList)

		if err != nil {
			logger.Log.Error("error while getting author profiles", zap.Error(err))
			result <- nil
			return
		}

		authorProfiles := funk.Map(resp.Profiles, func(profile *authPb.UserProfileProto) *socialPb.SocialProfile {
			return &socialPb.SocialProfile{
				Name:       profile.Name,
				PhotoUrl:   profile.PhotoUrl,
				Occupation: "farmer",
				UserId:     profile.LoginId,
			}
		}).([]*socialPb.SocialProfile)

		result <- authorProfiles
	}()

	return result
}

func GetSocialProfile(grpcContext context.Context, userId string) chan *socialPb.SocialProfile {
	result := make(chan *socialPb.SocialProfile)

	go func() {
		conn := auth_client.getConnection()
		if conn == nil {
			result <- nil
			return
		}

		client := authPb.NewProfileClient(conn)

		ctx := prepareCallContext(grpcContext)
		if ctx == nil {
			result <- nil
			return
		}

		resp, err := client.GetProfileById(
			ctx,
			&authPb.IdRequest{UserId: userId})
		if err != nil {
			logger.Error("Failed getting user profile", zap.Error(err))
			result <- nil
			return
		}

		result <- &socialPb.SocialProfile{
			Name:       resp.Name,
			PhotoUrl:   resp.PhotoUrl,
			Occupation: "farmer",
			UserId:     resp.LoginId,
		}
	}()

	return result
}

func IsUserAdmin(grpcContext context.Context, userId string) chan bool {
	result := make(chan bool)

	go func() {
		conn := auth_client.getConnection()
		if conn == nil {
			result <- false
			return
		}

		client := authPb.NewProfileClient(conn)

		ctx := prepareCallContext(grpcContext)
		if ctx == nil {
			result <- false
			return
		}

		resp, err := client.IsUserAdmin(ctx, &authPb.IdRequest{UserId: userId})
		if err != nil {
			logger.Error("Failed getting user profile", zap.Error(err))
			result <- false
			return
		}

		result <- resp.IsAdmin
	}()

	return result
}

func prepareCallContext(grpcContext context.Context) context.Context {
	jwtToken, err := grpc_auth.AuthFromMD(grpcContext, "bearer")
	if err != nil {
		logger.Error("Failed getting jwt token", zap.Error(err))
		return nil
	}

	return metadata.AppendToOutgoingContext(context.Background(), "Authorization", "bearer "+jwtToken)
}
