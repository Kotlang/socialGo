package models

import (
	"github.com/google/uuid"
)

type EventModel struct {
	EventId        string       `bson:"_id"`
	Title          string       `bson:"title"`
	Post           string       `bson:"post"`
	Type           string       `bson:"type"`
	StartAt        int64        `bson:"startAt"`
	EndAt          int64        `bson:"endAt"`
	OnlineLink     string       `bson:"onlineLink"`
	Description    string       `bson:"description"`
	MediaUrls      []MediaUrl   `bson:"mediaUrls"`
	WebPreviews    []WebPreview `bson:"webPreviews"`
	UserId         string       `bson:"userId"`
	ReferenceEvent string       `bson:"referenceEvent"`
	NumLikes       int          `bson:"numLikes"`
	NumShares      int          `bson:"numShares"`
	NumReplies     int          `bson:"numReplies"`
	NumAttendees   int64        `bson:"numAttendees"`
	NumSlots       int64        `bson:"numSlots"`
	Location       Location     `bson:"location"`
	Language       string       `bson:"language"`
	CreatedOn      int64        `bson:"createdOn"`
}

type EventLikeModel struct {
	UserId   string `bson:"userId"`
	PostId   string `bson:"postId"`
	PostType string `bson:"postType"`
}

func (p *EventLikeModel) Id() string {
	return p.UserId + "/" + p.PostId
}

func (m *EventModel) Id() string {
	if len(m.EventId) == 0 {
		m.EventId = uuid.NewString()
	}
	return m.EventId
}
