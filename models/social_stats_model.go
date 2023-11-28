package models

type SocialStatsModel struct {
	UserId    string `json:"userId" bson:"userId"`
	Followers int32  `json:"followers" bson:"followers"`
	Following int32  `json:"following" bson:"following"`
	Posts     int32  `json:"posts" bson:"posts"`
	Events    int32  `json:"events" bson:"events"`
}

func (p *SocialStatsModel) Id() string {
	return p.UserId
}
