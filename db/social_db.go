package db

import (
	"reflect"

	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type SocialDb struct{}

func (db *SocialDb) FeedPost(tenant string) *FeedPostRepository {
	repo := odm.AbstractRepository{
		CollectionName: "feed_post_" + tenant,
		Model:          reflect.TypeOf(models.FeedPostModel{}),
	}
	return &FeedPostRepository{repo}
}

func (db *SocialDb) Tag(tenant string) *TagRepository {
	repo := odm.AbstractRepository{
		CollectionName: "tag_" + tenant,
		Model:          reflect.TypeOf(models.PostTagModel{}),
	}
	return &TagRepository{repo}
}
