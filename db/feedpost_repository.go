package db

import (
	"time"

	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type FeedPostRepository struct {
	odm.AbstractRepository[models.FeedPostModel]
}

func (r *FeedPostRepository) GetEventFeed(
	eventStatus pb.EventStatus,
	postIds []string,
	referencePost string,
	pageNumber, pageSize int64) []models.FeedPostModel {

	now := time.Now().Unix()
	filters := bson.M{}
	filters["postType"] = pb.PostType_SOCIAL_EVENT.String()

	if len(postIds) > 0 {
		filters["_id"] = bson.M{"$in": postIds}
	}

	// parent post referencePost field is always empty string in db.
	filters["referencePost"] = referencePost

	if pb.EventStatus_PAST == eventStatus {
		filters["socialEventMetadata.endAt"] = bson.M{"$lt": now}
	} else if pb.EventStatus_ONGOING == eventStatus {
		filters["socialEventMetadata.startAt"] = bson.M{"$lt": now}
		filters["socialEventMetadata.endAt"] = bson.M{"$gt": now}
	} else if pb.EventStatus_FUTURE == eventStatus {
		filters["socialEventMetadata.startAt"] = bson.M{"$gt": now}
	}

	sort := bson.D{
		{Key: "createdOn", Value: -1},
		{Key: "numShares", Value: -1},
		{Key: "numReplies", Value: -1},
		{Key: "numLikes", Value: -1},
	}

	skip := pageNumber * pageSize
	resultChan, errChan := r.Find(filters, sort, pageSize, skip)

	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		logger.Error("Failed getting feed", zap.Error(err))
		return []models.FeedPostModel{}
	}
}

func (r *FeedPostRepository) GetFeed(
	feedFilters *pb.FeedFilters,
	referencePost string,
	pageNumber, pageSize int64) []models.FeedPostModel {

	filters := bson.M{}

	if feedFilters != nil {
		filters["postType"] = feedFilters.PostType.String()
	}

	if feedFilters != nil && len(feedFilters.Tag) > 0 {
		filters["tags"] = feedFilters.Tag
	}

	if feedFilters != nil && len(feedFilters.UserId) > 0 {
		filters["userId"] = feedFilters.UserId
	}

	// parent post referencePost field is always empty string in db.
	filters["referencePost"] = referencePost

	sort := bson.D{
		{Key: "createdOn", Value: -1},
		{Key: "numShares", Value: -1},
		{Key: "numReplies", Value: -1},
		{Key: "numLikes", Value: -1},
	}

	skip := pageNumber * pageSize
	resultChan, errChan := r.Find(filters, sort, pageSize, skip)

	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		logger.Error("Failed getting feed", zap.Error(err))
		return []models.FeedPostModel{}
	}
}
