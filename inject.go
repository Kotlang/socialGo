package main

import (
	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/service"
)

type Inject struct {
	SocialDb *db.SocialDb

	FeedPostService      *service.FeedpostService
	PostActionsService   *service.PostActionsService
	FollowGraphService   *service.FollowGraphService
	SocialStatsService   *service.SocialStatsService
	ContentMasterService *service.ContentMasterService
}

func NewInject() *Inject {
	inj := &Inject{}
	inj.SocialDb = &db.SocialDb{}

	inj.FeedPostService = service.NewFeedpostService(inj.SocialDb)
	inj.PostActionsService = service.NewPostActionsService(inj.SocialDb)
	inj.FollowGraphService = service.NewFollowGraphService(inj.SocialDb)
	inj.SocialStatsService = service.NewSocialStatsService(inj.SocialDb)
	inj.ContentMasterService = service.NewContentMasterService(inj.SocialDb)
	return inj
}
