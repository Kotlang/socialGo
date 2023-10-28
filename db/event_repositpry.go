package db

import (
	"time"

	pb "github.com/Kotlang/socialGo/generated"
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
	eventStatus pb.EventStatus,
	eventIds []string,
	referenceEvent string,
	pageNumber, pageSize int64) []models.EventModel {

	now := time.Now().Unix()
	filters := bson.M{}

	if len(eventIds) > 0 {
		filters["_id"] = bson.M{"$in": eventIds}
	}

	// parent event referenceEvent field is always empty string in db.
	filters["referenceEvent"] = referenceEvent

	if pb.EventStatus_PAST == eventStatus {
		filters["endAt"] = bson.M{"$lt": now}
	} else if pb.EventStatus_ONGOING == eventStatus {
		filters["startAt"] = bson.M{"$lt": now}
		filters["endAt"] = bson.M{"$gt": now}
	} else if pb.EventStatus_FUTURE == eventStatus {
		filters["startAt"] = bson.M{"$gt": now}
	}

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

func (r *EventRepository) GetComments(
	referenceEvent string,
	pageNumber, pageSize int64) []models.EventModel {
	filters := bson.M{"referenceEvent": referenceEvent}

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
		logger.Error("Failed getting comments", zap.Error(err))
		return []models.EventModel{}
	}
}
