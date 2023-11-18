package db

import (
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/odm"
)

type SocialDbInterface interface {
	FeedPost(tenant string) FeedPostRepositoryInterface
	Event(tenant string) EventRepositoryInterface
	Tag(tenant string) TagRepositoryInterface
	Comment(tenant string) CommentRepositoryInterface
	EventSubscribe(tenant string) EventSubscribeRepositoryInterface
	React(tenant string) ReactRepositoryInterface
	FollowersList(tenant string) FollowersListRepositoryInterface
	SocialStats(tenant string) SocialStatsRepositoryInterface
}

type SocialDb struct{}

func (db *SocialDb) FeedPost(tenant string) FeedPostRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.FeedPostModel]{
		Database:       tenant + "_social",
		CollectionName: "feed_post",
	}

	return &FeedPostRepository{repo}
}

func (db *SocialDb) Event(tenant string) EventRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.EventModel]{
		Database:       tenant + "_social",
		CollectionName: "feed_event",
	}

	return &EventRepository{repo}
}

func (db *SocialDb) Tag(tenant string) TagRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.PostTagModel]{
		Database:       tenant + "_social",
		CollectionName: "tag",
	}

	return &TagRepository{repo}
}
func (db *SocialDb) Comment(tenant string) CommentRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.CommentModel]{
		Database:       tenant + "_social",
		CollectionName: "comments",
	}
	return &CommentRepository{repo}
}

func (db *SocialDb) EventSubscribe(tenant string) EventSubscribeRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.EventSubscribeModel]{
		Database:       tenant + "_social",
		CollectionName: "event_subscribe",
	}
	return &EventSubscribeRepository{repo}
}

func (db *SocialDb) React(tenant string) ReactRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.ReactionModel]{
		Database:       tenant + "_social",
		CollectionName: "reaction",
	}

	return &ReactRepository{repo}
}

func (db *SocialDb) FollowersList(tenant string) FollowersListRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.FollowersListModel]{
		Database:       tenant + "_social",
		CollectionName: "followers_list",
	}

	return &FollowersListRepository{repo}
}

func (db *SocialDb) SocialStats(tenant string) SocialStatsRepositoryInterface {
	repo := odm.UnimplementedBootRepository[models.SocialStatsModel]{
		Database:       tenant + "_social",
		CollectionName: "social_stats",
	}

	return &SocialStatsRepository{repo}
}
