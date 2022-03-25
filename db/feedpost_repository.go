package db

import (
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
	postType string,
	tagFilter string,
	referencePost string,
	pageNumber, pageSize int64) []models.FeedPostModel {

	filters := bson.M{
		"postType": postType,
	}

	if len(tagFilter) > 0 {
		filters["tags"] = tagFilter
	}

	if len(referencePost) > 0 {
		filters["referencePost"] = referencePost
	}

	sort := bson.D{
		{"createdOn", -1},
		{"numShares", -1},
		{"numReplies", -1},
		{"numLikes", -1},
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
