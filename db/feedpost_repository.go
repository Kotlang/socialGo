package db

import (
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type FeedPostRepositoryInterface interface {
	odm.BootRepository[models.FeedPostModel]
	GetFeed(feedFilters *pb.FeedFilters, pageNumber, pageSize int64) []models.FeedPostModel
}
type FeedPostRepository struct {
	odm.UnimplementedBootRepository[models.FeedPostModel]
}

func (r *FeedPostRepository) GetFeed(
	feedFilters *socialPb.FeedFilters,
	pageNumber, pageSize int64) []models.FeedPostModel {

	filters := bson.M{}
	if feedFilters != nil {
		filters["postType"] = feedFilters.PostType.String()
	}
	if feedFilters != nil && len(feedFilters.ContentType) > 0 {
		filters["contentType"] = bson.M{"$in": feedFilters.ContentType}
	}

	if feedFilters != nil && len(feedFilters.Tag) > 0 {
		filters["tags"] = feedFilters.Tag
	}

	if feedFilters != nil && len(feedFilters.CreatedBy) > 0 {
		filters["userId"] = feedFilters.CreatedBy
	}

	filters["isDeleted"] = false

	sort := bson.D{
		{Key: "createdOn", Value: -1},
		{Key: "numShares", Value: -1},
		{Key: "numReplies", Value: -1},
		{Key: "numReacts", Value: -1},
	}

	skip := pageNumber * pageSize

	//Run a aggregation query to get the posts liked by the user as posts and likes are in different collections
	if feedFilters.FetchUserReactedPosts {

		filters["res.userId"] = feedFilters.UserId
		pipeline := bson.A{
			bson.D{
				{Key: "$lookup",
					Value: bson.D{
						{Key: "from", Value: "likes"},
						{Key: "localField", Value: "_id"},
						{Key: "foreignField", Value: "postId"},
						{Key: "as", Value: "res"},
					},
				},
			},
			bson.D{{Key: "$match", Value: filters}},
			bson.D{{Key: "$sort", Value: sort}},
			bson.D{{Key: "$skip", Value: skip}},
			bson.D{{Key: "$limit", Value: pageSize}},
		}

		resultsChan, errChan := r.Aggregate(pipeline)
		select {
		case res := <-resultsChan:
			return res
		case err := <-errChan:
			logger.Error("Failed getting feed", zap.Error(err))
			return []models.FeedPostModel{}
		}
	}

	if feedFilters.FetchUserCommentedPosts {
		logger.Info("Implement comment fetch logic")
		//TODO : Add the logic to fetch the posts commented by the user once the comments are seprated from the posts collection
	}

	//If like and comment fetch is not required then fetch the posts as usual
	resultChan, errChan := r.Find(filters, sort, pageSize, skip)

	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		logger.Error("Failed getting feed", zap.Error(err))
		return []models.FeedPostModel{}
	}
}
