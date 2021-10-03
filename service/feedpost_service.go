package service

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	s3client "github.com/Kotlang/socialGo/s3Client"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func (s *FeedpostService) saveTags(tenant string, tags []string) {
	for _, tag := range tags {
		existingTagRes := <-s.db.Tag(tenant).FindOneById(tag)

		var existingTag *models.PostTagModel
		// Need if-else since if existingTagRes.Value pointer is null, it cannot be typecasted.
		if existingTagRes.Value == nil {
			existingTag = &models.PostTagModel{
				Tag: tag,
			}
		} else {
			existingTag = existingTagRes.Value.(*models.PostTagModel)
		}

		existingTag.NumPosts += 1
		<-s.db.Tag(tenant).Save(existingTag)
	}
}

func (s *FeedpostService) CreatePost(ctx context.Context, req *pb.UserPostRequest) (*pb.UserPostProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	feedPostModel := s.db.FeedPost(tenant).GetModel(req).(*models.FeedPostModel)
	feedPostModel.UserId = userId
	feedPostModel.PostType = pb.UserPostRequest_PostType_name[int32(req.PostType)]

	// save post
	savePostAsync := s.db.FeedPost(tenant).Save(feedPostModel)

	// save tags
	s.saveTags(tenant, req.Tags)
	<-savePostAsync

	res := &pb.UserPostProto{}
	copier.Copy(res, feedPostModel)

	return res, nil
}

// Adds additional userProfile data, comments/answers to feedPost parameter.
func (s *FeedpostService) attachPostUserInfoAsync(
	grpcContext context.Context,
	feedPost *pb.UserPostProto,
	userId, tenant, userType string, attachAnswers bool) chan bool {
	done := make(chan bool)

	go func() {
		feedPost.HasFeedUserLiked = s.db.PostLike(tenant).IsExistsById(
			(&models.PostLikeModel{UserId: userId, PostId: feedPost.PostId}).Id(),
		)
		// get post author profile
		authorProfile := db.GetAuthorProfile(feedPost.UserId, grpcContext)
		if authorProfile != nil {
			feedPost.AuthorInfo = &pb.AuthorInfo{
				Name:       authorProfile.Name,
				PhotoUrl:   authorProfile.PhotoUrl,
				Occupation: "farmer",
			}
		}

		if attachAnswers {
			answers := s.db.FeedPost(tenant).GetFeed(
				pb.UserPostRequest_QnA_ANSWER.String(),
				"",
				feedPost.PostId,
				int64(0),
				int64(10))
			answersProto := []*pb.UserPostProto{}
			copier.Copy(&answersProto, answers)
			feedPost.AnswersThread = answersProto
		}

		done <- true
	}()

	return done
}

func (s *FeedpostService) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.UserPostProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	post := <-s.db.FeedPost(tenant).FindOneById(req.PostId)

	if post.Err != nil {
		return nil, status.Error(codes.Internal, post.Err.Error())
	}

	postProto := pb.UserPostProto{}
	copier.Copy(&postProto, post.Value)

	<-s.attachPostUserInfoAsync(ctx, &postProto, userId, tenant, "default", true)
	return &postProto, nil
}

func (s *FeedpostService) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.FeedResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	logger.Info("Getting feed for ", zap.String("feedType", req.PostType.String()))

	var tagFilter string
	if req.Filters != nil {
		tagFilter = req.Filters.Tag
	}

	feed := s.db.FeedPost(tenant).GetFeed(
		req.PostType.String(),
		tagFilter,
		"",
		int64(req.PageNumber),
		int64(req.PageSize))

	feedProto := []*pb.UserPostProto{}
	copier.Copy(&feedProto, feed)

	response := &pb.FeedResponse{Posts: feedProto}

	attachAnswers := (req.PostType == pb.GetFeedRequest_QnA_QUESTION)
	addUserPostActionsPromises := funk.Map(response.Posts, func(x *pb.UserPostProto) chan bool {
		return s.attachPostUserInfoAsync(ctx, x, userId, tenant, "default", attachAnswers)
	}).([]chan bool)
	for _, promise := range addUserPostActionsPromises {
		<-promise
	}

	return response, nil
}

func (s *FeedpostService) GetTags(ctx context.Context, req *pb.GetTagsRequest) (*pb.TagListResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)
	tags := s.db.Tag(tenant).GetTagsRanked()

	return &pb.TagListResponse{
		Tag: funk.Map(tags, func(x models.PostTagModel) string { return x.Tag }).([]string),
	}, nil
}

func (s *FeedpostService) GetMediaUploadUrl(ctx context.Context, req *pb.MediaUploadRequest) (*pb.MediaUploadURL, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	uploadUrl, downloadUrl := s3client.GetPresignedUrlForPosts(tenant, userId, req.MediaExtension)
	return &pb.MediaUploadURL{
		UploadUrl: uploadUrl,
		MediaUrl:  downloadUrl,
	}, nil
}
