package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type EventSubscribeRepositoryInterface interface {
	odm.BootRepository[models.EventSubscribeModel]
	IsSubscriber(userId string, eventId string) bool
}

type EventSubscribeRepository struct {
	odm.UnimplementedBootRepository[models.EventSubscribeModel]
}

func (r *EventSubscribeRepository) IsSubscriber(userId string, eventId string) bool {
	return r.IsExistsById(userId + "/" + eventId)

}
