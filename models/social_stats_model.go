package models

type SocialStatsModel struct {
	UserId    string `json:"userId" bson:"userId"`
	Followers int32  `json:"followers" bson:"followers"`
	Following int32  `json:"following" bson:"following"`
	Posts     int32  `json:"posts" bson:"posts"`
	Events    int32  `json:"events" bson:"events"`
}

func (p *SocialStatsModel) Id() string {
	return GetSocialStatsId(p.UserId)
}

// returns the social stats id for the given user
func GetSocialStatsId(userId string) string {
	return userId
}
