package models

type EventSubscribeModel struct {
	EventSubscribeId string `bson:"_id"`
	UserId           string `bson:"userId"`
	EventId          string `bson:"eventId"`
	CreatedOn        int64  `bson:"createdOn"`
}

func (m *EventSubscribeModel) Id() string {
	if len(m.EventSubscribeId) == 0 {
		m.EventSubscribeId = GetEventSubscribeId(m.UserId, m.EventId)
	}
	return m.EventSubscribeId
}

// returns the event subscribe id for the given user and event
func GetEventSubscribeId(userId, eventId string) string {
	return userId + "/" + eventId
}
