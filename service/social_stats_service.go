package service

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/SaiNageswarS/go-api-boot/auth"
)

type SocialStatsService struct {
	socialPb.UnimplementedSocialStatsServer
	db db.SocialDbInterface
}

func NewSocialStatsService(db db.SocialDbInterface) *SocialStatsService {
	return &SocialStatsService{
		db: db,
	}
}

func (s *SocialStatsService) GetStats(ctx context.Context, req *socialPb.GetStatsRequest) (*socialPb.GetStatsResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	if len(req.UserId) > 0 {
		userId = req.UserId
	}

	stats := s.db.SocialStats(tenant).GetStats(userId)

	return &socialPb.GetStatsResponse{
		PostsCount:     stats.Posts,
		FollowersCount: stats.Followers,
		FollowingCount: stats.Following,
	}, nil
}
