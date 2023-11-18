package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type EventSubscribeRepositoryInterface interface {
	odm.BootRepository[models.EventSubscribeModel]
}

type EventSubscribeRepository struct {
	odm.UnimplementedBootRepository[models.EventSubscribeModel]
}
