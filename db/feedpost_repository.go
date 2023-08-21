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

	if feedFilters != nil {
		now := time.Now().Unix()

		if pb.EventStatus_PAST == feedFilters.EventStatus {
			filters["socialEventMetadata.endat"] = bson.M{"$lt": now}
		} else if pb.EventStatus_ONGOING == feedFilters.EventStatus {
			filters["socialEventMetadata.startat"] = bson.M{"$lt": now}
			filters["socialEventMetadata.endat"] = bson.M{"$gt": now}
		} else if pb.EventStatus_FUTURE == feedFilters.EventStatus {
			filters["socialEventMetadata.startat"] = bson.M{"$gt": now}
		}
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
