package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type SocialStatsRepositoryInterface interface {
	odm.BootRepository[models.SocialStatsModel]
	GetStats(userId string) *models.SocialStatsModel
	UpdatePostCount(userId string, posts int32) chan error
	UpdateFollowerCount(userId string, followers int32) chan error
	UpdateFollowingCount(userId string, following int32) chan error
}

type SocialStatsRepository struct {
	odm.UnimplementedBootRepository[models.SocialStatsModel]
}

func (r *SocialStatsRepository) GetStats(userId string) *models.SocialStatsModel {
	currentStatsChan, errChan := r.FindOneById(userId)

	select {
	case currentStats := <-currentStatsChan:
		return currentStats
	case <-errChan:
		return &models.SocialStatsModel{
			UserId: userId,
		}
	}
}

func (r *SocialStatsRepository) UpdatePostCount(userId string, posts int32) chan error {
	currentStats := r.GetStats(userId)
	currentStats.Posts += posts
	return r.Save(currentStats)
}

func (r *SocialStatsRepository) UpdateFollowerCount(userId string, followers int32) chan error {
	currentStats := r.GetStats(userId)
	currentStats.Followers += followers
	return r.Save(currentStats)
}

func (r *SocialStatsRepository) UpdateFollowingCount(userId string, following int32) chan error {
	currentStats := r.GetStats(userId)
	currentStats.Following += following
	return r.Save(currentStats)
}
