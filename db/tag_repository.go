package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type TagRepositoryInterface interface {
	odm.BootRepository[models.PostTagModel]
	FindTagsRanked() []models.PostTagModel
}

type TagRepository struct {
	odm.UnimplementedBootRepository[models.PostTagModel]
}

func (r *TagRepository) FindTagsRanked() []models.PostTagModel {
	filters := bson.M{}
	sort := bson.D{
		{Key: "numPosts", Value: -1},
	}
	resultChan, errChan := r.Find(filters, sort, 10, 0)

	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		logger.Error("Failed getting tags", zap.Error(err))
		return []models.PostTagModel{}
	}
}
