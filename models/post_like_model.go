package models

type PostLikeModel struct {
	UserId string `bson:"userId"`
	PostId string `bson:"postId"`
}

func (p *PostLikeModel) Id() string {
	return p.UserId + "/" + p.PostId
}
