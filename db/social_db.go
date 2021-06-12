package db

import (
	"reflect"

	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type SocialDb struct{}

func (db *SocialDb) FeedPost(tenant string) *FeedPostRepository {
	repo := odm.AbstractRepository{
		CollectionName: "login_" + tenant,
		Model:          reflect.TypeOf(models.FeedPostModel{}),
	}
	return &FeedPostRepository{repo}
}
