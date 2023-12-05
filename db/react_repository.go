package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/SaiNageswarS/go-api-boot/odm"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type ReactRepository struct {
	odm.AbstractRepository[models.ReactionModel]
}

// TODO: Use mongo find one with projection to get only reaction field instead of fetching whole document.
func (r *ReactRepository) GetUserReactions(entityId, userId string) []string {
	var reactions []string
	reactionResChan, errorResChan := r.FindOneById(r.GetId(entityId, userId))

	select {
	case reactionRes := <-reactionResChan:
		reactions = reactionRes.Reaction
		return reactions
	case err := <-errorResChan:
		if err == mongo.ErrNoDocuments {
			return []string{}
		} else {
			logger.Error("Error while fetching reactions", zap.Error(err))
		}
		return nil
	}
}
func (r *ReactRepository) GetId(entityId, userId string) string {
	return userId + "/" + entityId
}
