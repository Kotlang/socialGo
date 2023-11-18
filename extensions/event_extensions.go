package extensions

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

// Adds additional userProfile data, comments/answers to feedEvent parameter.
func AttachEventInfoAsync(
	socialDb db.SocialDbInterface,
	grpcContext context.Context,
	feedEvent *pb.EventProto,
	userId, tenant, userType string) chan bool {

	// logger.Info("AttachPostUserInfoAsync", zap.Any("feedEvent", feedEvent))

	done := make(chan bool)

	go func() {
		feedEvent.FeedUserReactions = socialDb.React(tenant).GetUserReactions(feedEvent.EventId, userId)
		done <- true
	}()

	return done
}

func GetSubscribedPostIds(db db.SocialDbInterface, tenant string, subscriberId string) chan []string {
	postIds := make(chan []string)

	go func() {
		subscribeFilters := bson.M{}
		subscribeFilters["userId"] = subscriberId

		subscribeCountChan, errChan := db.EventSubscribe(tenant).CountDocuments(subscribeFilters)
		var count int64 = 0
		select {
		case subscribeCount := <-subscribeCountChan:
			count = subscribeCount
		case err := <-errChan:
			logger.Error("Failed getting subscribed post count", zap.Error(err))
			postIds <- []string{}
			return
		}

		subscribeEventChan, errChan := db.EventSubscribe(tenant).Find(subscribeFilters, bson.D{}, count, 0)
		subscribedPostIds := []string{}
		select {
		case subscribedPosts := <-subscribeEventChan:
			for _, subscribedPosts := range subscribedPosts {
				subscribedPostIds = append(subscribedPostIds, subscribedPosts.EventId)
			}
			postIds <- subscribedPostIds
		case err := <-errChan:
			logger.Error("Failed getting subscribed events", zap.Error(err))
			postIds <- []string{}
			return
		}
	}()

	return postIds
}
