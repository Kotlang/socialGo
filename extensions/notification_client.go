package extensions

import (
	"context"
	"errors"
	"os"
	"sync"

	notificationPb "github.com/Kotlang/socialGo/generated/notification"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var notification_client *NotificationClient = &NotificationClient{}

type NotificationClient struct {
	cached_conn        *grpc.ClientConn
	conn_creation_lock sync.Mutex
}

func (c *NotificationClient) getNotificationConnection() *grpc.ClientConn {
	c.conn_creation_lock.Lock()
	defer c.conn_creation_lock.Unlock()

	if c.cached_conn == nil || c.cached_conn.GetState().String() != "READY" {
		if val, ok := os.LookupEnv("NOTIFICATION_TARGET"); ok {
			conn, err := grpc.Dial(val, grpc.WithInsecure(), grpc.WithBlock())
			if err != nil {
				logger.Error("Failed getting connection with notification service", zap.Error(err))
				return nil
			}
			c.cached_conn = conn
		} else {
			logger.Error("Failed to get NOTIFICATION_TARGET env variable")
		}

	}

	return c.cached_conn
}

func RegisterEvent(grpcContext context.Context, event *notificationPb.RegisterEventRequest) chan error {
	errChan := make(chan error)

	go func() {
		// call notification service.
		conn := notification_client.getNotificationConnection()
		if conn == nil {
			errChan <- errors.New("failed to get connection with notification service")
			return
		}

		client := notificationPb.NewNotificationServiceClient(conn)

		ctx := prepareCallContext(grpcContext)
		if ctx == nil {
			errChan <- errors.New("failed to get context")
			return
		}

		_, err := client.RegisterEvent(ctx, event)
		if err != nil {
			errChan <- err
			return
		}
		errChan <- nil
	}()

	return errChan
}
