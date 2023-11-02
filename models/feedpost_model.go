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

type SocialEventMetadata struct {
	Name         string   `bson:"name"`
	Type         string   `bson:"type"`
	StartAt      int64    `bson:"startAt"`
	EndAt        int64    `bson:"endAt"`
	OnlineLink   string   `bson:"onlineLink"`
	Description  string   `bson:"description"`
	NumAttendees int64    `bson:"numAttendees"`
	NumSlots     int64    `bson:"numSlots"`
	Location     Location `bson:"location"`
}

type FeedPostModel struct {
	PostId              string               `bson:"_id"`
	Title               string               `bson:"title"`
	Post                string               `bson:"post"`
	MediaUrls           []MediaUrl           `bson:"mediaUrls"`
	WebPreviews         []WebPreview         `bson:"webPreviews"`
	ReferencePost       string               `bson:"referencePost"`
	Replies             []string             `bson:"replies"`
	PostType            string               `bson:"postType"`
	ContentType         []string             `bson:"contentType"`
	UserId              string               `bson:"userId"`
	NumLikes            int                  `bson:"numLikes"`
	NumShares           int                  `bson:"numShares"`
	NumReplies          int                  `bson:"numReplies"`
	Tags                []string             `bson:"tags"`
	CreatedOn           int64                `bson:"createdOn"`
	Language            string               `bson:"language"`
	SocialEventMetadata *SocialEventMetadata `bson:"socialEventMetadata"`
}

func (m *FeedPostModel) Id() string {
	if len(m.PostId) == 0 {
		m.PostId = uuid.NewString()
	}
	return m.PostId
}
