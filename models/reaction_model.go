package models

type ReactionModel struct {
	UserId     string   `bson:"userId"`
	EntityId   string   `bson:"entityId"`
	Reaction   []string `bson:"reaction"`
	ReactionOn string   `bson:"reactOn"`
}

func (p *ReactionModel) Id() string {
	return GetReactionId(p.UserId, p.EntityId)
}

// returns the reaction id for the given user and entity
func GetReactionId(userId, entityId string) string {
	return userId + "/" + entityId
}
