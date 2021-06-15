package service

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PostActionsService struct {
	pb.UnimplementedPostActionsServer
	db *db.SocialDb
}

func NewPostActionsService(db *db.SocialDb) *PostActionsService {
	return &PostActionsService{
		db: db,
	}
}

func (s *PostActionsService) LikePost(ctx context.Context, req *pb.PostIdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	postLikeModel := &models.PostLikeModel{
		PostId: req.PostId,
		UserId: userId,
	}

	isLikeExists := s.db.PostLike(tenant).IsExistsById(postLikeModel.Id())

	if !isLikeExists {
		likeAsyncSaveReq := s.db.PostLike(tenant).Save(postLikeModel)
		feedPostRes := <-s.db.FeedPost(tenant).FindOneById(req.PostId)

		if feedPostRes.Err != nil {
			logger.Error("Probably post is not found", zap.Error(feedPostRes.Err))
			return nil, status.Error(codes.NotFound, "Post not found")
		}
		feedPost := feedPostRes.Value.(*models.FeedPostModel)
		feedPost.NumLikes += 1
		<-s.db.FeedPost(tenant).Save(feedPost)
		<-likeAsyncSaveReq
	}

	return &pb.StatusResponse{Status: "success"}, nil
}

func (s *PostActionsService) UnLikePost(ctx context.Context, req *pb.PostIdRequest) (*pb.StatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	postLikeModel := &models.PostLikeModel{
		PostId: req.PostId,
		UserId: userId,
	}

	isLikeExists := s.db.PostLike(tenant).IsExistsById(postLikeModel.Id())

	if isLikeExists {
		unLikeAsyncSaveReq := s.db.PostLike(tenant).DeleteById(postLikeModel.Id())
		feedPostRes := <-s.db.FeedPost(tenant).FindOneById(req.PostId)

		if feedPostRes.Err != nil {
			logger.Error("Probably post is not found", zap.Error(feedPostRes.Err))
			return nil, status.Error(codes.NotFound, "Post not found")
		}
		feedPost := feedPostRes.Value.(*models.FeedPostModel)
		feedPost.NumLikes -= 1
		<-s.db.FeedPost(tenant).Save(feedPost)
		<-unLikeAsyncSaveReq
	}

	return &pb.StatusResponse{Status: "success"}, nil
}
