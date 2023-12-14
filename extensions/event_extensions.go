package extensions

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
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
		feedEvent.HasFeedUserSubscribed = socialDb.EventSubscribe(tenant).IsSubscriber(userId, feedEvent.EventId)
		if feedEvent.AuthorInfo != nil && feedEvent.AuthorInfo.UserId != "" {
			authorProfile := <-GetSocialProfile(grpcContext, feedEvent.AuthorInfo.UserId)
			feedEvent.AuthorInfo = authorProfile
		}
		done <- true
	}()

	return done
}

// AttachMultipleEventInfoAsync attaches event reaction info to multiple event proto.
func AttachMultipleEventInfoAsync(
	socialDb db.SocialDbInterface,
	grpcContext context.Context,
	feedEvents []*socialPb.EventProto,
	userId, tenant, userType string) chan bool {

	done := make(chan bool)

	entityId := []string{}
	authorId := []string{}
	for _, feedEvent := range feedEvents {
		entityId = append(entityId, models.GetReactionId(feedEvent.EventId, userId))
		if feedEvent.AuthorInfo != nil && feedEvent.AuthorInfo.UserId != "" {
			authorId = append(authorId, feedEvent.AuthorInfo.UserId)
		}
	}

	go func() {

		filter := bson.M{
			"_id": bson.M{
				"$in": entityId,
			},
		}

		//get user reactions for the events
		reactionResChan, errChan := socialDb.React(tenant).Find(filter, bson.D{}, 0, 0)
		//get user subscribers for the events
		subscriberResChan, subscriberErrChan := socialDb.EventSubscribe(tenant).Find(filter, bson.D{}, 0, 0)
		//get author profiles for the events
		authorResChan := GetSocialProfiles(grpcContext, authorId)

		select {
		case reactions := <-reactionResChan:
			//map of event id to user reactions
			reactionMap := make(map[string][]string)
			for _, reaction := range reactions {
				reactionMap[reaction.EntityId] = reaction.Reaction
			}
			//attach user reactions to the events
			for _, feedEvent := range feedEvents {
				feedEvent.FeedUserReactions = reactionMap[feedEvent.EventId]
				if feedEvent.FeedUserReactions == nil {
					feedEvent.FeedUserReactions = []string{}
				}
			}
		case err := <-errChan:
			if err != nil && err != mongo.ErrNoDocuments {
				logger.Error("Error while fetching reactions", zap.Error(err))
			}
			logger.Info("No reactions found")
		}

		select {
		case subscribers := <-subscriberResChan:
			//map of event id to user subscribed status
			subscriberMap := make(map[string]bool)
			for _, subscriber := range subscribers {
				subscriberMap[subscriber.EventId] = true
			}
			//attach user subscribed status to the events
			for _, feedEvent := range feedEvents {
				feedEvent.HasFeedUserSubscribed = subscriberMap[feedEvent.EventId]
			}

		case err := <-subscriberErrChan:
			if err != nil && err != mongo.ErrNoDocuments {
				logger.Error("Error while fetching subscribers", zap.Error(err))
			}
			logger.Info("No subscribers found")
		}

		authorProfiles := <-authorResChan
		//map of user id to author profile
		authorMap := make(map[string]*socialPb.SocialProfile)
		for _, authorProfile := range authorProfiles {
			authorMap[authorProfile.UserId] = authorProfile
		}
		//attach author profile to the events
		for _, feedEvent := range feedEvents {
			if authorMap[feedEvent.AuthorInfo.UserId] != nil {
				feedEvent.AuthorInfo = authorMap[feedEvent.AuthorInfo.UserId]
			}
		}

		done <- true
	}()
	return done
}

// GetSubscribedEventIds returns the list of subscribed event ids for the given user
func GetSubscribedEventIds(db db.SocialDbInterface, tenant string, subscriberId string) chan []string {
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
