package extensions

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// AttachEventInfoAsync attaches event reaction info to the event proto.
func AttachEventInfoAsync(
	socialDb db.SocialDbInterface,
	grpcContext context.Context,
	feedEvent *socialPb.EventProto,
	userId, tenant, userType string) chan bool {

	done := make(chan bool)

	go func() {
		feedEvent.FeedUserReactions = socialDb.React(tenant).GetUserReactions(feedEvent.EventId, userId)
		done <- true
	}()

	return done
}

// AttachMultipleEventInfoAsync attaches event reaction info to multiple event proto.
func AttachMultipleEventInfoAsync(
	socialDb *db.SocialDb,
	grpcContext context.Context,
	feedEvents []*socialPb.EventProto,
	userId, tenant, userType string) chan bool {

	done := make(chan bool)

	eventIds := funk.Map(feedEvents, func(feedEvent *socialPb.EventProto) string {
		return userId + "/" + feedEvent.EventId
	}).([]string)

	go func() {
		filter := bson.M{
			"_id": bson.M{
				"$in": eventIds,
			},
		}

		reactionResChan, errChan := socialDb.React(tenant).Find(filter, bson.D{}, 0, 0)

		select {
		case reactions := <-reactionResChan:
			for _, reaction := range reactions {
				for _, feedEvent := range feedEvents {
					if feedEvent.EventId == reaction.EntityId {
						feedEvent.FeedUserReactions = reaction.Reaction
					}
				}
			}
		case err := <-errChan:
			if err != nil && err != mongo.ErrNoDocuments {
				logger.Error("Error while fetching reactions", zap.Error(err))
			}
			logger.Info("No reactions found")
		}
		done <- true
	}()
	return done
}

// GetSubscribedEventIds returns the list of subscribed event ids for the given user
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
