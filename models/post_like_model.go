package models

type PostLikeModel struct {
	UserId   string `bson:"userId"`
	PostId   string `bson:"postId"`
	PostType string `bson:"postType"`
}

func (p *PostLikeModel) Id() string {
	return p.UserId + "/" + p.PostId
}
