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

func (s *PostActionsService) LikePost(ctx context.Context, req *pb.PostIdRequest) (*pb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	postLikeModel := &models.PostLikeModel{
		PostId: req.PostId,
		UserId: userId,
	}

	isLikeExists := s.db.PostLike(tenant).IsExistsById(postLikeModel.Id())

	if !isLikeExists {
		feedPostChan, errChan := s.db.FeedPost(tenant).FindOneById(req.PostId)
		select {
		case feedPost := <-feedPostChan:
			postLikeModel.PostType = feedPost.PostType
			likeAsyncSaveReq := s.db.PostLike(tenant).Save(postLikeModel)
			<-likeAsyncSaveReq
			feedPost.NumLikes = feedPost.NumLikes + 1
			<-s.db.FeedPost(tenant).Save(feedPost)
		case err := <-errChan:
			logger.Error("Probably post is not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, "Post not found. "+err.Error())
		}
	}

	return &pb.SocialStatusResponse{Status: "success"}, nil
}

func (s *PostActionsService) UnLikePost(ctx context.Context, req *pb.PostIdRequest) (*pb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	postLikeModel := &models.PostLikeModel{
		PostId: req.PostId,
		UserId: userId,
	}

	isLikeExists := s.db.PostLike(tenant).IsExistsById(postLikeModel.Id())

	if isLikeExists {
		feedPostChan, errChan := s.db.FeedPost(tenant).FindOneById(req.PostId)
		select {
		case feedPost := <-feedPostChan:
			unLikeAsyncSaveReq := s.db.PostLike(tenant).DeleteById(postLikeModel.Id())
			<-unLikeAsyncSaveReq
			feedPost.NumLikes = feedPost.NumLikes - 1
			<-s.db.FeedPost(tenant).Save(feedPost)
		case err := <-errChan:
			logger.Error("Probably post is not found", zap.Error(err))
			return nil, status.Error(codes.NotFound, "Post not found. "+err.Error())
		}
	}

	return &pb.SocialStatusResponse{Status: "success"}, nil
}
