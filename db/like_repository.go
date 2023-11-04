package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type LikeRepository struct {
	odm.AbstractRepository[models.PostLikeModel]
}
