package service

import (
	"context"
	"fmt"

	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/extensions"
	notificationPb "github.com/Kotlang/socialGo/generated/notification"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FollowGraphService struct {
	socialPb.UnimplementedFollowGraphServer
	db db.SocialDbInterface
}

func NewFollowGraphService(db db.SocialDbInterface) *FollowGraphService {
	return &FollowGraphService{
		db: db,
	}
}

func (s *FollowGraphService) FollowUser(ctx context.Context, req *socialPb.FollowUserRequest) (*socialPb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// save follower. userId is the follower.
	followerModel := &models.FollowersListModel{
		UserId:     req.UserId,
		FollowerId: userId,
	}
	// fetch follower profile.
	followerIdResChan := extensions.GetSocialProfile(ctx, userId)

	// check if follow relationship already exists.
	if s.db.FollowersList(tenant).IsExistsById(followerModel.Id()) {
		return &socialPb.SocialStatusResponse{Status: "ALREADY_FOLLOWING"}, nil
	}

	err := <-s.db.FollowersList(tenant).Save(followerModel)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	folower := <-followerIdResChan
	// TODO: update title text to contain district name
	err = <-extensions.RegisterEvent(ctx, &notificationPb.RegisterEventRequest{
		EventType: "user.follow",
		Title:     fmt.Sprintf("%s आपसे नवाचार के माध्यम से सीखना चाहते है", folower.Name),
		Body:      fmt.Sprintf("%s नवाचार पर आपके साझा किये हुए अनुभव को उपयोगी मानते है", folower.Name),
		TemplateParameters: map[string]string{
			"follower": userId,
			"followee": req.UserId,
		},
		Topic:       fmt.Sprintf("%s.user.follow", tenant),
		TargetUsers: []string{req.UserId},
	})

	if err != nil {
		logger.Error("Failed to register event", zap.Error(err))
	}

	followerCountPromise := s.db.SocialStats(tenant).UpdateFollowerCount(req.UserId, 1)
	followsCountPromise := s.db.SocialStats(tenant).UpdateFollowingCount(userId, 1)

	<-followerCountPromise
	<-followsCountPromise

	return &socialPb.SocialStatusResponse{Status: "success"}, nil
}

func (s *FollowGraphService) UnfollowUser(ctx context.Context, req *socialPb.UnFollowUserRequest) (*socialPb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// save follower. userId is the follower.
	followerModel := &models.FollowersListModel{
		UserId:     req.UserId,
		FollowerId: userId,
	}

	// check if follow relationship already exists.
	if !s.db.FollowersList(tenant).IsExistsById(followerModel.Id()) {
		return &socialPb.SocialStatusResponse{Status: "NOT_FOLLOWING"}, nil
	}

	err := <-s.db.FollowersList(tenant).DeleteById(followerModel.Id())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	followerCountPromise := s.db.SocialStats(tenant).UpdateFollowerCount(req.UserId, -1)
	followsCountPromise := s.db.SocialStats(tenant).UpdateFollowingCount(userId, -1)

	<-followerCountPromise
	<-followsCountPromise

	return &socialPb.SocialStatusResponse{Status: "success"}, nil
}

func (s *FollowGraphService) GetFollowers(ctx context.Context, req *socialPb.GetFollowersRequest) (*socialPb.GetFollowersResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	followerIds := s.db.FollowersList(tenant).GetFollowers(userId, int64(req.PageNumber), int64(req.PageSize))
	followers := <-extensions.GetSocialProfiles(ctx, followerIds)

	return &socialPb.GetFollowersResponse{Followers: followers}, nil
}

func (s *FollowGraphService) GetFollowing(ctx context.Context, req *socialPb.GetFollowingRequest) (*socialPb.GetFollowingResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	followingIds := s.db.FollowersList(tenant).GetFollowing(userId, int64(req.PageNumber), int64(req.PageSize))
	following := <-extensions.GetSocialProfiles(ctx, followingIds)

	return &socialPb.GetFollowingResponse{Following: following}, nil
}

func (s *FollowGraphService) IsFollowing(ctx context.Context, req *socialPb.IsFollowingRequest) (*socialPb.IsFollowingResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	followerModel := &models.FollowersListModel{
		UserId:     req.Followee,
		FollowerId: req.Follower,
	}

	isExists := s.db.FollowersList(tenant).IsExistsById(followerModel.Id())
	return &socialPb.IsFollowingResponse{IsFollowing: isExists}, nil
}
