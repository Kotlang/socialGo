package extensions

import (
	"context"
	"fmt"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	"github.com/jinzhu/copier"
)

// Adds additional userProfile data, comments/answers to feedEvent parameter.
func AttachEventInfoAsync(
	socialDb *db.SocialDb,
	grpcContext context.Context,
	feedEvent *pb.EventProto,
	userId, tenant, userType string,
	attachAnswers bool) chan bool {

	// logger.Info("AttachPostUserInfoAsync", zap.Any("feedEvent", feedEvent))

	done := make(chan bool)

	go func() {
		feedEvent.HasFeedUserLiked = socialDb.PostLike(tenant).IsExistsById(
			(&models.EventLikeModel{UserId: userId, PostId: feedEvent.EventId}).Id(),
		)

		// get post author profile
		authorProfile := <-GetSocialProfile(grpcContext, feedEvent.UserId)
		feedEvent.AuthorInfo = authorProfile
		if attachAnswers {
			answers := socialDb.Event(tenant).GetComments(
				feedEvent.EventId,
				int64(0),
				int64(10))
			answersProto := []*pb.EventProto{}
			copier.Copy(&answersProto, answers)
			feedEvent.AnswersThread = answersProto

			// recursively attach authorInfo to answers.
			for _, answerProto := range feedEvent.AnswersThread {
				fmt.Println("answerProto\n", answerProto)
				<-AttachEventInfoAsync(socialDb, grpcContext, answerProto, userId, tenant, userType, false)
			}
		}

		done <- true
	}()

	return done
}
