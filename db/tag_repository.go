package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type TagRepository struct {
	odm.AbstractRepository
}

func (r *TagRepository) GetTagsRanked() []models.PostTagModel {
	filters := bson.M{}
	sort := bson.D{
		{"numPosts", -1},
	}
	res := <-r.Find(filters, sort, 10, 0)
	if res.Err != nil {
		logger.Error("Failed fetching tags", zap.Error(res.Err))
		return []models.PostTagModel{}
	}
	return res.Value.([]models.PostTagModel)
}
