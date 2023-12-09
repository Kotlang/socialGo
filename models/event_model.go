package models

import (
	"github.com/google/uuid"
)

type EventModel struct {
	EventId      string           `bson:"_id"`
	Title        string           `bson:"title"`
	Type         string           `bson:"type"`
	StartAt      int64            `bson:"startAt"`
	EndAt        int64            `bson:"endAt"`
	OnlineLink   string           `bson:"onlineLink"`
	Description  string           `bson:"description"`
	MediaUrls    []MediaUrl       `bson:"mediaUrls"`
	WebPreviews  []WebPreview     `bson:"webPreviews"`
	UserId       string           `bson:"userId"`
	NumReacts    map[string]int64 `bson:"numReacts"`
	NumShares    int64            `bson:"numShares"`
	NumReplies   int64            `bson:"numReplies"`
	NumAttendees int32            `bson:"numAttendees"`
	NumSlots     int32            `bson:"numSlots"`
	Location     Location         `bson:"location"`
	Language     string           `bson:"language"`
	CreatedOn    int64            `bson:"createdOn"`
	IsDeleted    bool             `bson:"isDeleted"`
}

func (m *EventModel) Id() string {
	if len(m.EventId) == 0 {
		m.EventId = uuid.NewString()
	}
	return m.EventId
}
