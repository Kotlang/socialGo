package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type CommentRepositoryInterface interface {
	odm.BootRepository[models.CommentModel]
	GetComments(parentId string, pageNumber, pageSize int64) []models.CommentModel
}
type CommentRepository struct {
	odm.UnimplementedBootRepository[models.CommentModel]
}

func (c *CommentRepository) GetComments(parentId string, pageNumber, pageSize int64) []models.CommentModel {
	filters := bson.M{}
	filters["parentId"] = parentId
	filters["isDeleted"] = false

	sort := bson.D{
		{Key: "createdOn", Value: -1},
		{Key: "numShares", Value: -1},
		{Key: "numReplies", Value: -1},
		{Key: "numLikes", Value: -1},
	}
	skip := pageNumber * pageSize

	resultChan, errChan := c.Find(filters, sort, pageSize, skip)

	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		logger.Error("Failed getting comments", zap.Error(err))
		return []models.CommentModel{}
	}
}
