package service

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/SaiNageswarS/go-api-boot/auth"
)

type SocialStatsService struct {
	pb.UnimplementedSocialStatsServer
	db *db.SocialDb
}

func NewSocialStatsService(db *db.SocialDb) *SocialStatsService {
	return &SocialStatsService{
		db: db,
	}
}

func (s *SocialStatsService) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	stats := s.db.SocialStats(tenant).GetStats(userId)

	return &pb.GetStatsResponse{
		PostsCount:     stats.Posts,
		FollowersCount: stats.Followers,
		FollowingCount: stats.Following,
	}, nil
}
