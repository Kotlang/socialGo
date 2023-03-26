package models

type PostTagModel struct {
	Tag            string `bson:"tag"`
	TagDescription string `bson:"tagDescription"`
	NumPosts       int    `bson:"numPosts"`
	CreatedOn      int64  `bson:"createdOn"`
}

func (t *PostTagModel) Id() string {
	return t.Tag
}
