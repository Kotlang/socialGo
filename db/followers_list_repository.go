package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type FollowersListRepository struct {
	odm.AbstractRepository[models.FollowersListModel]
}

func (r *FollowersListRepository) GetFollowers(userId string, pageNumber, pageSize int64) []string {
	skip := pageNumber * pageSize

	followersChan, errChan := r.Find(bson.M{
		"userId": userId,
	}, bson.D{{"createdOn", -1}}, pageSize, skip)

	select {
	case followers := <-followersChan:
		return funk.Map(followers, func(follower models.FollowersListModel) string {
			return follower.FollowerId
		}).([]string)
	case err := <-errChan:
		logger.Error("Failed getting followers", zap.Error(err))
		return nil
	}
}

func (r *FollowersListRepository) GetFollowing(userId string, pageNumber, pageSize int64) []string {
	skip := pageNumber * pageSize

	followingChan, errChan := r.Find(bson.M{
		"followerId": userId,
	}, bson.D{{"createdOn", -1}}, pageSize, skip)

	select {
	case following := <-followingChan:
		return funk.Map(following, func(following models.FollowersListModel) string {
			return following.UserId
		}).([]string)
	case err := <-errChan:
		logger.Error("Failed getting following", zap.Error(err))
		return nil
	}
}
