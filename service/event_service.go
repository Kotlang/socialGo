package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/extensions"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EventService struct {
	pb.UnimplementedEventsServer
	db db.SocialDbInterface
}

func NewEventService(db db.SocialDbInterface) *EventService {
	return &EventService{
		db: db,
	}
}

func (s *EventService) CreateEvent(ctx context.Context, req *pb.CreateEventRequest) (*pb.EventProto, error) {
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

	//TODO: Add field event to socialStatsModel and increment eventsCount
	savePostCountPromise := s.db.SocialStats(tenant).UpdatePostCount(userId, 1)

	// wait for async operations to finish.
	<-saveEventPromise
	<-saveTagsPromise
	<-savePostCountPromise

	eventModelChan, errChan := s.db.Event(tenant).FindOneById(eventModel.EventId)

	select {
	case eventModel := <-eventModelChan:
		res := &pb.EventProto{}
		copier.Copy(res, eventModel)

		//populate hasUsserLiked field
		attachAuthorInfoPromise := extensions.AttachEventInfoAsync(s.db, ctx, res, userId, tenant, "default")

		err := <-extensions.RegisterEvent(ctx, &pb.RegisterEventRequest{
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

		<-attachAuthorInfoPromise
		return res, nil
	case err := <-errChan:
		return nil, status.Error(codes.Internal, err.Error())
	}
}

// TODO: Add feild isDeleted to eventModel and update it to true
func (s *EventService) DeleteEvent(ctx context.Context, req *pb.EventIdRequest) (*pb.SocialStatusResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	eventModelChan, errChan := s.db.Event(tenant).FindOneById(req.EventId)
	select {
	case eventModel := <-eventModelChan:
		eventModel.IsDeleted = true
		<-s.db.Event(tenant).Save(eventModel)
	case err := <-errChan:
		logger.Error("Failed to delete event", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	//TODO: Add field event to socialStatsModel and decrement eventsCount
	//TODO: Delete all the comments and likes on the event

	return &pb.SocialStatusResponse{Status: "success"}, nil
}

func (s *EventService) GetEvent(ctx context.Context, req *pb.EventIdRequest) (*pb.EventProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	EventProto := pb.EventProto{}

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

func (s *EventService) GetEventFeed(ctx context.Context, req *pb.GetEventFeedRequest) (*pb.EventFeedResponse, error) {

	userId, tenant := auth.GetUserIdAndTenant(ctx)
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	logger.Info("Getting feed for ", zap.String("feedType", req.Filters.EventStatus.String()))

	eventStatus := pb.EventStatus_FUTURE
	eventIds := []string{}
	if req.Filters != nil {
		if req.Filters.GetSubscribedEvents {
			eventIds = <-extensions.GetSubscribedPostIds(s.db, tenant, userId)
			if len(eventIds) == 0 {
				return &pb.EventFeedResponse{Events: []*pb.EventProto{}}, nil
			}
		}
		eventStatus = req.Filters.EventStatus
	}

	feed := s.db.Event(tenant).GetEventFeed(
		eventStatus,
		eventIds,
		int64(req.PageNumber),
		int64(req.PageSize))

	feedProto := []*pb.EventProto{}
	copier.Copy(&feedProto, feed)
	response := &pb.EventFeedResponse{Events: feedProto}
	response.PageNumber = req.PageNumber
	response.PageSize = req.PageSize

	addUserPostActionsPromises := funk.Map(response.Events, func(x *pb.EventProto) chan bool {
		return extensions.AttachEventInfoAsync(s.db, ctx, x, userId, tenant, "default")
	}).([]chan bool)
	for _, promise := range addUserPostActionsPromises {
		<-promise
	}

	return response, nil
}

func (s *EventService) SubscribeEvent(ctx context.Context, req *pb.EventIdRequest) (*pb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	eventSubscribeModel := s.db.EventSubscribe(tenant).GetModel(req)
	eventSubscribeModel.UserId = userId

	isExistsById := s.db.EventSubscribe(tenant).IsExistsById(eventSubscribeModel.Id())

	if isExistsById {
		return &pb.SocialStatusResponse{Status: "success"}, nil
	}

	err := <-s.db.EventSubscribe(tenant).Save(eventSubscribeModel)

	if err != nil {
		logger.Error("Failed to subscribe event", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())

	}
	return &pb.SocialStatusResponse{Status: "success"}, nil
}
