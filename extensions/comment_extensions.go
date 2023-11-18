package extensions

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
)

func AttachCommentUserInfoAsync(
	socialDb db.SocialDbInterface,
	grpcContext context.Context,
	comment *pb.CommentProto,
	userId, tenant, userType string) chan bool {

	// logger.Info("AttachPostUserInfoAsync", zap.Any("feedPost", feedPost))

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
