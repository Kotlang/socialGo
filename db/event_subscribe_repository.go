package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type EventSubscribeRepositoryInterface interface {
	odm.BootRepository[models.EventSubscribeModel]
	IsSubscriber(userId string, eventId string) bool
	FetchEventSubscriberList(eventId string) []models.EventSubscribeModel
}

type EventSubscribeRepository struct {
	odm.UnimplementedBootRepository[models.EventSubscribeModel]
}

func (r *EventSubscribeRepository) IsSubscriber(userId string, eventId string) bool {
	return r.IsExistsById(models.GetEventSubscribeId(userId, eventId))
}

func (r *EventSubscribeRepository) FetchEventSubscriberList(eventId string) []models.EventSubscribeModel {
	filter := bson.M{"eventId": eventId}

	subscriberCountResChan, errResChan := r.CountDocuments(filter)
	var subscriberCount int64
	select {
	case subscriberCount = <-subscriberCountResChan:

	case err := <-errResChan:
		logger.Error("Failed getting count of subcriber", zap.Error(err))
	}
	eventSubscribeResChan, errResChan := r.Find(filter, nil, subscriberCount, 0)
	select {
	case eventSubscribeList := <-eventSubscribeResChan:
		return eventSubscribeList
	case err := <-errResChan:
		logger.Error("Failed getting subscriber list", zap.Error(err))
		return []models.EventSubscribeModel{}
	}
}
