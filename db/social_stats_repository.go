package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type SocialStatsRepository struct {
	odm.AbstractRepository[models.SocialStatsModel]
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

func (r *SocialStatsRepository) UpdateEventCount(userId string, events int32) chan error {
	currentStats := r.GetStats(userId)
	currentStats.Events += events
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
