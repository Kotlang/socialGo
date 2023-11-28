package db

import (
	socialPb "github.com/Kotlang/socialGo/generated/social"
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
	feedFilters *socialPb.FeedFilters,
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

	filters["isDeleted"] = false

	sort := bson.D{
		{Key: "createdOn", Value: -1},
		{Key: "numShares", Value: -1},
		{Key: "numReplies", Value: -1},
		{Key: "numReacts", Value: -1},
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
