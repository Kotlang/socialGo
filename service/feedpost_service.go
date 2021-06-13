package service

import (
	"context"
	"time"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	s3client "github.com/Kotlang/socialGo/s3Client"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"github.com/thoas/go-funk"
	"go.uber.org/zap"
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
				Tag:       tag,
				CreatedOn: time.Now().Unix(),
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
	feedPostModel.CreatedOn = time.Now().Unix()

	// save post
	savePostAsync := s.db.FeedPost(tenant).Save(feedPostModel)

	// save tags
	s.saveTags(tenant, req.Tags)
	<-savePostAsync

	res := &pb.UserPostProto{}
	copier.Copy(res, feedPostModel)

	return res, nil
}

func (s *FeedpostService) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.FeedResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	logger.Info("Getting feed for ", zap.String("feedType", req.PostType.String()))

	tagFilters := []string{}
	if req.Filters != nil {
		tagFilters = req.Filters.Tags
	}

	feed := s.db.FeedPost(tenant).GetFeed(req.PostType.String(), tagFilters, int64(req.PageNumber), int64(req.PageSize))

	feedProto := []*pb.UserPostProto{}

	copier.Copy(&feedProto, feed)
	logger.Info("Got feed as ", zap.Any("feed", feedProto))

	return &pb.FeedResponse{Posts: feedProto}, nil
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
