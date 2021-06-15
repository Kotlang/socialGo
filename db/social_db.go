package db

import (
	"reflect"

	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type SocialDb struct{}

func (db *SocialDb) FeedPost(tenant string) *FeedPostRepository {
	repo := odm.AbstractRepository{
		Database:       tenant + "_social",
		CollectionName: "feed_post",
		Model:          reflect.TypeOf(models.FeedPostModel{}),
	}
	return &FeedPostRepository{repo}
}

func (db *SocialDb) Tag(tenant string) *TagRepository {
	repo := odm.AbstractRepository{
		Database:       tenant + "_social",
		CollectionName: "tag",
		Model:          reflect.TypeOf(models.PostTagModel{}),
	}
	return &TagRepository{repo}
}

func (db *SocialDb) PostLike(tenant string) *PostLikeRepository {
	repo := odm.AbstractRepository{
		Database:       tenant + "_social",
		CollectionName: "likes",
		Model:          reflect.TypeOf(models.PostLikeModel{}),
	}
	return &PostLikeRepository{repo}
}
