package models

import (
	"github.com/google/uuid"
)

type CommentModel struct {
	CommentId   string           `bson:"_id"`
	Content     string           `bson:"content"`
	MediaUrls   []MediaUrl       `bson:"mediaUrls"`
	WebPreviews []WebPreview     `bson:"webPreviews"`
	ParentId    string           `bson:"parentId"`
	UserId      string           `bson:"userId"`
	CommentOn   string           `bson:"commentOn"`
	NumReacts   map[string]int64 `bson:"numReacts"`
	NumReplies  int              `bson:"numReplies"`
	NumShares   int              `bson:"numShares"`
	CreatedOn   int64            `bson:"createdOn"`
	IsDeleted   bool             `bson:"isDeleted"`
}

func (c *CommentModel) Id() string {
	if len(c.CommentId) == 0 {
		c.CommentId = uuid.NewString()
	}
	return c.CommentId
}
