package db

import (
	"time"

	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type EventRepository struct {
	odm.AbstractRepository[models.EventModel]
}

func (r *EventRepository) GetEventFeed(
	eventStatus socialPb.EventStatus,
	eventIds []string,
	pageNumber, pageSize int64) []models.EventModel {

	now := time.Now().Unix()
	filters := bson.M{}

	if len(eventIds) > 0 {
		filters["_id"] = bson.M{"$in": eventIds}
	}

	if socialPb.EventStatus_PAST == eventStatus {
		filters["endAt"] = bson.M{"$lt": now}
	} else if socialPb.EventStatus_ONGOING == eventStatus {
		filters["startAt"] = bson.M{"$lt": now}
		filters["endAt"] = bson.M{"$gt": now}
	} else if socialPb.EventStatus_FUTURE == eventStatus {
		filters["startAt"] = bson.M{"$gt": now}
	}
	filters["isDeleted"] = false

	sort := bson.D{
		{Key: "createdOn", Value: -1},
		{Key: "numShares", Value: -1},
		{Key: "numReplies", Value: -1},
		{Key: "numLikes", Value: -1},
	}

	skip := pageNumber * pageSize
	resultChan, errChan := r.Find(filters, sort, pageSize, skip)

	select {
	case res := <-resultChan:
		return res
	case err := <-errChan:
		logger.Error("Failed getting feed", zap.Error(err))
		return []models.EventModel{}
	}
}
