package service

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/extensions"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FollowGraphService struct {
	pb.UnimplementedFollowGraphServer
	db *db.SocialDb
}

func NewFollowGraphService(db *db.SocialDb) *FollowGraphService {
	return &FollowGraphService{
		db: db,
	}
}

func (s *FollowGraphService) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// save follower. userId is the follower.
	followerModel := &models.FollowersListModel{
		UserId:     req.UserId,
		FollowerId: userId,
	}

	// check if follow relationship already exists.
	if s.db.FollowersList(tenant).IsExistsById(followerModel.Id()) {
		return &pb.SocialStatusResponse{Status: "ALREADY_FOLLOWING"}, nil
	}

	err := <-s.db.FollowersList(tenant).Save(followerModel)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	followerCountPromise := s.db.SocialStats(tenant).UpdateFollowerCount(req.UserId, 1)
	followsCountPromise := s.db.SocialStats(tenant).UpdateFollowingCount(userId, 1)

	<-followerCountPromise
	<-followsCountPromise

	return &pb.SocialStatusResponse{Status: "success"}, nil
}

func (s *FollowGraphService) UnfollowUser(ctx context.Context, req *pb.UnFollowUserRequest) (*pb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// save follower. userId is the follower.
	followerModel := &models.FollowersListModel{
		UserId:     req.UserId,
		FollowerId: userId,
	}

	// check if follow relationship already exists.
	if !s.db.FollowersList(tenant).IsExistsById(followerModel.Id()) {
		return &pb.SocialStatusResponse{Status: "NOT_FOLLOWING"}, nil
	}

	err := <-s.db.FollowersList(tenant).DeleteById(followerModel.Id())

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	followerCountPromise := s.db.SocialStats(tenant).UpdateFollowerCount(req.UserId, -1)
	followsCountPromise := s.db.SocialStats(tenant).UpdateFollowingCount(userId, -1)

	<-followerCountPromise
	<-followsCountPromise

	return &pb.SocialStatusResponse{Status: "success"}, nil
}

func (s *FollowGraphService) GetFollowers(ctx context.Context, req *pb.GetFollowersRequest) (*pb.GetFollowersResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	followerIds := s.db.FollowersList(tenant).GetFollowers(userId, int64(req.PageNumber), int64(req.PageSize))
	followers := <-extensions.GetSocialProfiles(ctx, followerIds)

	return &pb.GetFollowersResponse{Followers: followers}, nil
}

func (s *FollowGraphService) GetFollowing(ctx context.Context, req *pb.GetFollowingRequest) (*pb.GetFollowingResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	if req.PageSize == 0 {
		req.PageSize = 10
	}

	followingIds := s.db.FollowersList(tenant).GetFollowing(userId, int64(req.PageNumber), int64(req.PageSize))
	following := <-extensions.GetSocialProfiles(ctx, followingIds)

	return &pb.GetFollowingResponse{Following: following}, nil
}
