package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type EventSubscribeRepository struct {
	odm.AbstractRepository[models.EventSubscribeModel]
}
