package models

import (
	"github.com/google/uuid"
)

type MediaUrl struct {
	Url      string `bson:"url"`
	MimeType string `bson:"mimeType"`
}

type WebPreview struct {
	Title        string `bson:"title"`
	PreviewImage string `bson:"previewImage"`
	Url          string `bson:"url"`
	Description  string `bson:"description"`
}

type FeedPostModel struct {
	PostId        string       `bson:"_id"`
	Title         string       `bson:"title"`
	Post          string       `bson:"post"`
	MediaUrls     []MediaUrl   `bson:"mediaUrls"`
	WebPreviews   []WebPreview `bson:"webPreviews"`
	ReferencePost string       `bson:"referencePost"`
	Replies       []string     `bson:"replies"`
	PostType      string       `bson:"postType"`
	UserId        string       `bson:"userId"`
	NumLikes      int          `bson:"numLikes"`
	NumShares     int          `bson:"numShares"`
	NumReplies    int          `bson:"numReplies"`
	Tags          []string     `bson:"tags"`
	CreatedOn     int64        `bson:"createdOn"`
	Language      string       `bson:"language"`
}

func (m *FeedPostModel) Id() string {
	if len(m.PostId) == 0 {
		m.PostId = uuid.NewString()
	}
	return m.PostId
}
