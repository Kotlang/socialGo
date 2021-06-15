package main

import (
	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/service"
	"github.com/joho/godotenv"
)

type Inject struct {
	SocialDb *db.SocialDb

	FeedPostService    *service.FeedpostService
	PostActionsService *service.PostActionsService
}

func NewInject() *Inject {
	godotenv.Load()
	inj := &Inject{}

	inj.SocialDb = &db.SocialDb{}

	inj.FeedPostService = service.NewFeedpostService(inj.SocialDb)
	inj.PostActionsService = service.NewPostActionsService(inj.SocialDb)
	return inj
}
