package models

type ReactionModel struct {
	UserId     string   `bson:"userId"`
	EntityId   string   `bson:"entityId"`
	Reaction   []string `bson:"reaction"`
	ReactionOn string   `bson:"reactOn"`
}

func (p *ReactionModel) Id() string {
	return p.UserId + "/" + p.EntityId
}
