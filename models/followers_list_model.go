package models

type FollowersListModel struct {
	UserId     string `bson:"userId"`
	FollowerId string `bson:"followerId"`

	// added automatically by go-api-boot
	CreatedOn int64 `bson:"createdOn"`
}

func (p *FollowersListModel) Id() string {
	return GetFollowersListId(p.UserId, p.FollowerId)
}

// returns the followers list id for the given user and follower
func GetFollowersListId(userId, followerId string) string {
	return userId + "/" + followerId
}
