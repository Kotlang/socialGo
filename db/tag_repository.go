package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type TagRepository struct {
	odm.AbstractRepository[models.PostTagModel]
}

func (r *TagRepository) FindByLanguage(language string) []models.PostTagModel {
	filters := bson.M{"language": language}
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
