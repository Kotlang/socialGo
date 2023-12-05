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

type Location struct {
	Lat  float64 `bson:"lat"`
	Long float64 `bson:"long"`
}

type FeedPostModel struct {
	PostId      string           `bson:"_id"`
	Title       string           `bson:"title"`
	Post        string           `bson:"post"`
	MediaUrls   []MediaUrl       `bson:"mediaUrls"`
	WebPreviews []WebPreview     `bson:"webPreviews"`
	PostType    string           `bson:"postType"`
	UserId      string           `bson:"userId"`
	NumReacts   map[string]int64 `bson:"numReacts"`
	NumShares   int64            `bson:"numShares"`
	NumReplies  int64            `bson:"numReplies"`
	Tags        []string         `bson:"tags"`
	CreatedOn   int64            `bson:"createdOn"`
	Language    string           `bson:"language"`
	IsDeleted   bool             `bson:"isDeleted"`
}

func (m *FeedPostModel) Id() string {
	if len(m.PostId) == 0 {
		m.PostId = uuid.NewString()
	}
	return m.PostId
}
