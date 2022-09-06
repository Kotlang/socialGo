package models

type FollowersListModel struct {
	UserId     string `bson:"userId"`
	FollowerId string `bson:"followerId"`

	// added automatically by go-api-boot
	CreatedOn int64 `bson:"createdOn"`
}

func (p *FollowersListModel) Id() string {
	return p.UserId + "/" + p.FollowerId
}
