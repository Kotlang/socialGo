package main

import (
	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/service"
	"github.com/SaiNageswarS/go-api-boot/cloud"
)

type Inject struct {
	SocialDb db.SocialDbInterface
	cloudFns cloud.Cloud

	FeedPostService    *service.FeedpostService
	ActionsService     *service.ActionsService
	FollowGraphService *service.FollowGraphService
	SocialStatsService *service.SocialStatsService
	EventService       *service.EventService
}

func NewInject() *Inject {
	inj := &Inject{}
	inj.SocialDb = &db.SocialDb{}
	inj.cloudFns = &cloud.GCP{}

	inj.FeedPostService = service.NewFeedpostService(inj.SocialDb, inj.cloudFns)
	inj.ActionsService = service.NewActionsService(inj.SocialDb)
	inj.FollowGraphService = service.NewFollowGraphService(inj.SocialDb)
	inj.SocialStatsService = service.NewSocialStatsService(inj.SocialDb)
	inj.EventService = service.NewEventService(inj.SocialDb)
	return inj
}
