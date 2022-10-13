package models

type PostTagModel struct {
	Language  string `bson:"language"`
	Tag       string `bson:"tag"`
	NumPosts  int    `bson:"numPosts"`
	CreatedOn int64  `bson:"createdOn"`
}

func (t *PostTagModel) Id() string {
	return t.Tag
}
