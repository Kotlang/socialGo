package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type PostLikeRepository struct {
	odm.AbstractRepository[models.PostLikeModel]
}
