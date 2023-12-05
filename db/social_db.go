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

func (db *SocialDb) Event(tenant string) *EventRepository {
	repo := odm.AbstractRepository[models.EventModel]{
		Database:       tenant + "_social",
		CollectionName: "feed_event",
	}

	return &EventRepository{repo}
}

func (db *SocialDb) Tag(tenant string) *TagRepository {
	repo := odm.AbstractRepository[models.PostTagModel]{
		Database:       tenant + "_social",
		CollectionName: "tag",
	}

	return &TagRepository{repo}
}
func (db *SocialDb) Comment(tenant string) *CommentRepository {
	repo := odm.AbstractRepository[models.CommentModel]{
		Database:       tenant + "_social",
		CollectionName: "comments",
	}
	return &CommentRepository{repo}
}

func (db *SocialDb) EventSubscribe(tenant string) *EventSubscribeRepository {
	repo := odm.AbstractRepository[models.EventSubscribeModel]{
		Database:       tenant + "_social",
		CollectionName: "event_subscribe",
	}
	return &EventSubscribeRepository{repo}
}

func (db *SocialDb) React(tenant string) *ReactRepository {
	repo := odm.AbstractRepository[models.ReactionModel]{
		Database:       tenant + "_social",
		CollectionName: "reaction",
	}

	return &ReactRepository{repo}
}

func (db *SocialDb) FollowersList(tenant string) *FollowersListRepository {
	repo := odm.AbstractRepository[models.FollowersListModel]{
		Database:       tenant + "_social",
		CollectionName: "followers_list",
	}

	return &FollowersListRepository{repo}
}

func (db *SocialDb) SocialStats(tenant string) *SocialStatsRepository {
	repo := odm.AbstractRepository[models.SocialStatsModel]{
		Database:       tenant + "_social",
		CollectionName: "social_stats",
	}

	return &SocialStatsRepository{repo}
}
