package models

import (
	"github.com/google/uuid"
)

type FeedPostModel struct {
	Title         string   `bson:"title"`
	PostText      string   `bson:"postText"`
	MediaUrls     []string `bson:"mediaUrls"`
	ReferencePost string   `bson:"referencePost"`
	Replies       []string `bson:"replies"`
	PostType      string   `bson:"postType"`
	UserId        string   `bson:"userId"`
	NumLikes      int      `bson:"numLikes"`
	NumShares     int      `bson:"numShares"`
	NumReplies    int      `bson:"NumReplies"`
}

func (m *FeedPostModel) Id() string {
	return uuid.NewString()
}
