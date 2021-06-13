package service

import (
	"context"
	"time"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type FeedpostService struct {
	pb.UnimplementedUserPostServer
	db *db.SocialDb
}

func NewFeedpostService(db *db.SocialDb) *FeedpostService {
	return &FeedpostService{
		db: db,
	}
}

func (s *FeedpostService) CreatePost(ctx context.Context, req *pb.UserPostRequest) (*pb.UserPostProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	feedPostModel := s.db.FeedPost(tenant).GetModel(req).(*models.FeedPostModel)
	feedPostModel.UserId = userId
	feedPostModel.PostType = pb.UserPostRequest_PostType_name[int32(req.PostType)]
	feedPostModel.CreatedOn = time.Now().Unix()

	<-s.db.FeedPost(tenant).Save(feedPostModel)
	res := &pb.UserPostProto{}
	copier.Copy(res, feedPostModel)

	return res, nil
}

func (s *FeedpostService) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.FeedResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	logger.Info("Getting feed for ", zap.String("feedType", req.PostType.String()))

	tagFilters := []string{}
	if req.Filters != nil {
		tagFilters = req.Filters.Tags
	}

	feed := s.db.FeedPost(tenant).GetFeed(req.PostType.String(), tagFilters, int64(req.PageNumber), int64(req.PageSize))

	feedProto := []*pb.UserPostProto{}

	copier.Copy(&feedProto, feed)
	logger.Info("Got feed as ", zap.Any("feed", feedProto))

	return &pb.FeedResponse{Posts: feedProto}, nil
}
