package extensions

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	socialPb "github.com/Kotlang/socialGo/generated/social"
)

func AttachCommentUserInfoAsync(
	socialDb *db.SocialDb,
	grpcContext context.Context,
	comment *socialPb.CommentProto,
	userId, tenant, userType string) chan bool {

	done := make(chan bool)

	go func() {
		comment.UserReactions = socialDb.React(tenant).GetUserReactions(comment.CommentId, userId)

		// get comment author profile
		authorProfile := <-GetSocialProfile(grpcContext, comment.UserId)
		comment.AuthorInfo = authorProfile

		done <- true
	}()

	return done
}
