package extensions

import (
	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	"github.com/jinzhu/copier"
)

func SaveTags(db *db.SocialDb, tenant string, tags []string) chan bool {
	savedTagsPromise := make(chan bool)

	go func() {
		for _, tag := range tags {
			existingTagChan, errChan := db.Tag(tenant).FindOneById(tag)
			select {
			case existingTag := <-existingTagChan:
				existingTag.NumPosts++
				<-db.Tag(tenant).Save(existingTag)
			case <-errChan:
				newTag := &models.PostTagModel{
					Tag:      tag,
					NumPosts: 1,
				}
				<-db.Tag(tenant).Save(newTag)
			}
		}

		savedTagsPromise <- true
	}()

	return savedTagsPromise
}

// Adds additional userProfile data, comments/answers to feedPost parameter.
func AttachPostUserInfoAsync(
	socialDb *db.SocialDb,
	authClient *AuthClient,
	feedPost *pb.UserPostProto,
	userId, tenant, userType string, attachAnswers bool) chan bool {
	done := make(chan bool)

	go func() {
		feedPost.HasFeedUserLiked = socialDb.PostLike(tenant).IsExistsById(
			(&models.PostLikeModel{UserId: userId, PostId: feedPost.PostId}).Id(),
		)
		// get post author profile
		authorProfile := <-authClient.GetAuthorProfile(feedPost.UserId)
		if authorProfile != nil {
			feedPost.AuthorInfo = &pb.UserProfile{
				Name:       authorProfile.Name,
				PhotoUrl:   authorProfile.PhotoUrl,
				Occupation: "farmer",
				UserId:     feedPost.UserId,
			}
		}

		if attachAnswers {
			answers := socialDb.FeedPost(tenant).GetFeed(
				pb.UserPostRequest_QnA_ANSWER.String(),
				&pb.FeedFilters{},
				feedPost.PostId,
				int64(0),
				int64(10))
			answersProto := []*pb.UserPostProto{}
			copier.Copy(&answersProto, answers)
			feedPost.AnswersThread = answersProto

			// recursively attach authorInfo to answers.
			for _, answerProto := range feedPost.AnswersThread {
				<-AttachPostUserInfoAsync(socialDb, authClient, answerProto, userId, tenant, userType, false)
			}
		}

		done <- true
	}()

	return done
}
