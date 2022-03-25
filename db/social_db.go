package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type SocialDb struct{}

func (db *SocialDb) FeedPost(tenant string) *FeedPostRepository {
	repo := odm.AbstractRepository[models.FeedPostModel]{
		Database:       tenant + "_social",
		CollectionName: "feed_post",
	}
	return &FeedPostRepository{repo}
}

func (db *SocialDb) Tag(tenant string) *TagRepository {
	repo := odm.AbstractRepository[models.PostTagModel]{
		Database:       tenant + "_social",
		CollectionName: "tag",
	}
	return &TagRepository{repo}
}

func (db *SocialDb) PostLike(tenant string) *PostLikeRepository {
	repo := odm.AbstractRepository[models.PostLikeModel]{
		Database:       tenant + "_social",
		CollectionName: "likes",
	}
	return &PostLikeRepository{repo}
}
