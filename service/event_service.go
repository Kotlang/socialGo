package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/extensions"
	notificationsPb "github.com/Kotlang/socialGo/generated/notification"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventService struct {
	socialPb.UnimplementedEventsServer
	db db.SocialDbInterface
}

func NewEventService(db db.SocialDbInterface) *EventService {
	return &EventService{
		db: db,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, req *socialPb.CreateEventRequest) (*socialPb.EventProto, error) {
	err := ValidateEventRequest(req)
	if err != nil {
		return nil, err
	}

	userId, tenant := auth.GetUserIdAndTenant(ctx)
	// map proto to model.
	eventModel := s.db.Event(tenant).GetModel(req)
	eventModel.UserId = userId

	if len(strings.TrimSpace(eventModel.Language)) == 0 {
		eventModel.Language = "english"
	}

	// save event
	saveEventPromise := s.db.Event(tenant).Save(eventModel)

	// save tags.
	saveTagsPromise := extensions.SaveTags(s.db, tenant, req.Tags)

	// update event count in social stats.
	saveEventCountPromise := s.db.SocialStats(tenant).UpdateEventCount(userId, 1)

	// wait for async operations to finish.
	<-saveEventPromise
	<-saveTagsPromise
	<-saveEventCountPromise

	eventModelChan, errChan := s.db.Event(tenant).FindOneById(eventModel.EventId)

	select {
	case eventModel := <-eventModelChan:
		res := &socialPb.EventProto{}
		copier.Copy(res, eventModel)

		//populate hasUserReacted field
		attachEventInfoPromise := extensions.AttachEventInfoAsync(s.db, ctx, res, userId, tenant, "default")

		err := <-extensions.RegisterEvent(ctx, &notificationsPb.RegisterEventRequest{
			EventType: "event.created",
			TemplateParameters: map[string]string{
				"postId": eventModel.EventId,
				"body":   eventModel.Description,
				"title":  eventModel.Title,
			},
			Topic:       fmt.Sprintf("%s.event.created", tenant),
			TargetUsers: []string{userId},
		})
		if err != nil {
			logger.Error("Failed to register event", zap.Error(err))
		}

		<-attachEventInfoPromise
		return res, nil
	case err := <-errChan:
		return nil, status.Error(codes.Internal, err.Error())
	}
}

func (s *EventService) DeleteEvent(ctx context.Context, req *socialPb.EventIdRequest) (*socialPb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	eventModelChan, errChan := s.db.Event(tenant).FindOneById(req.EventId)
	select {
	case eventModel := <-eventModelChan:
		eventModel.IsDeleted = true
		<-s.db.Event(tenant).Save(eventModel)
	case err := <-errChan:
		logger.Error("Failed to delete event", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// update event count in social stats.
	saveEventCountPromise := s.db.SocialStats(tenant).UpdateEventCount(userId, -1)
	<-saveEventCountPromise

	return &socialPb.SocialStatusResponse{Status: "success"}, nil
}

func (s *EventService) GetEvent(ctx context.Context, req *socialPb.EventIdRequest) (*socialPb.EventProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	EventProto := socialPb.EventProto{}

	filters := bson.M{}
	filters["_id"] = req.EventId
	filters["isDeleted"] = false

	eventChan, errChan := s.db.Event(tenant).FindOne(filters)
	select {
	case event := <-eventChan:
		copier.Copy(&EventProto, event)
	case err := <-errChan:
		if err == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, "Event not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	<-extensions.AttachEventInfoAsync(s.db, ctx, &EventProto, userId, tenant, "default")
	return &EventProto, nil
}

func (s *EventService) GetEventFeed(ctx context.Context, req *socialPb.GetEventFeedRequest) (*socialPb.EventFeedResponse, error) {

	userId, tenant := auth.GetUserIdAndTenant(ctx)
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	logger.Info("Getting feed for ", zap.String("feedType", req.Filters.EventStatus.String()))

	eventStatus := socialPb.EventStatus_FUTURE
	eventIds := []string{}
	if req.Filters != nil {
		if req.Filters.GetSubscribedEvents {
			eventIds = <-extensions.GetSubscribedEventIds(s.db, tenant, userId)
			if len(eventIds) == 0 {
				return &socialPb.EventFeedResponse{Events: []*socialPb.EventProto{}}, nil
			}
		}
		eventStatus = req.Filters.EventStatus
	}

	feed := s.db.Event(tenant).GetEventFeed(
		eventStatus,
		eventIds,
		int64(req.PageNumber),
		int64(req.PageSize))

	feedProto := []*socialPb.EventProto{}
	copier.Copy(&feedProto, feed)
	response := &socialPb.EventFeedResponse{Events: feedProto}
	response.PageNumber = req.PageNumber
	response.PageSize = req.PageSize

	attachEventInfoPromise := extensions.AttachMultipleEventInfoAsync(s.db, ctx, response.Events, userId, tenant, "default")
	<-attachEventInfoPromise

	return response, nil
}

func (s *EventService) SubscribeEvent(ctx context.Context, req *socialPb.EventIdRequest) (*socialPb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	eventSubscribeModel := s.db.EventSubscribe(tenant).GetModel(req)
	eventSubscribeModel.UserId = userId

	isExistsById := s.db.EventSubscribe(tenant).IsExistsById(eventSubscribeModel.Id())

	if isExistsById {
		return &socialPb.SocialStatusResponse{Status: "success"}, nil
	}

	err := <-s.db.EventSubscribe(tenant).Save(eventSubscribeModel)

	if err != nil {
		logger.Error("Failed to subscribe event", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())

	}
	return &socialPb.SocialStatusResponse{Status: "success"}, nil
}
