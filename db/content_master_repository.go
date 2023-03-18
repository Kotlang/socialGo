package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
)

type ContentMasterRepository struct {
	odm.AbstractRepository[models.ContentMasterModel]
}

func (p *ContentMasterRepository) FindByLanguage(language string) (chan []models.ContentMasterModel, chan error) {
	return p.Find(bson.M{"language": language}, nil, 0, 0)
}
