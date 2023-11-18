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
	socialDb *db.SocialDb,
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

func GetSubscribedEventIds(db *db.SocialDb, tenant string, subscriberId string) chan []string {
	eventIds := make(chan []string)

	go func() {
		subscribeFilters := bson.M{}
		subscribeFilters["userId"] = subscriberId

		subscribeCountChan, errChan := db.EventSubscribe(tenant).CountDocuments(subscribeFilters)
		var count int64 = 0
		select {
		case subscribeCount := <-subscribeCountChan:
			count = subscribeCount
		case err := <-errChan:
			logger.Error("Failed getting subscribed event count", zap.Error(err))
			eventIds <- []string{}
			return
		}

		subscribeEventChan, errChan := db.EventSubscribe(tenant).Find(subscribeFilters, bson.D{}, count, 0)
		subscribedeventIds := []string{}
		select {
		case subscribedevents := <-subscribeEventChan:
			for _, subscribedevents := range subscribedevents {
				subscribedeventIds = append(subscribedeventIds, subscribedevents.EventId)
			}
			eventIds <- subscribedeventIds
		case err := <-errChan:
			logger.Error("Failed getting subscribed events", zap.Error(err))
			eventIds <- []string{}
			return
		}
	}()

	return eventIds
}
