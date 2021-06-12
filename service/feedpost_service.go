package service

import (
	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
)

type FeedpostService struct {
	pb.UnimplementedUserPostServer
	db *db.SocialDb
}

func NewFeedpostService(db *db.SocialDb) *FeedpostService {
	return &FeedpostService{
		db: db,
	}
}
