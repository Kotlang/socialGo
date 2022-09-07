package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/extensions"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/Kotlang/socialGo/models"
	s3client "github.com/Kotlang/socialGo/s3Client"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/azure"
	"github.com/SaiNageswarS/go-api-boot/bootUtils"
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

func (s *FeedpostService) CreatePost(ctx context.Context, req *pb.UserPostRequest) (*pb.UserPostProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// map proto to model.
	feedPostModel := s.db.FeedPost(tenant).GetModel(req)
	feedPostModel.UserId = userId
	feedPostModel.PostType = pb.UserPostRequest_PostType_name[int32(req.PostType)]

	// save post.
	savePostPromise := s.db.FeedPost(tenant).Save(feedPostModel)

	// save tags.
	saveTagsPromise := extensions.SaveTags(s.db, tenant, req.Tags)

	// if it is a comment/answer increment numReplies
	if len(feedPostModel.ReferencePost) > 0 {
		parentPostChan, errChan := s.db.FeedPost(tenant).FindOneById(feedPostModel.ReferencePost)
		select {
		case parentPost := <-parentPostChan:
			parentPost.NumReplies = parentPost.NumReplies + 1
			<-s.db.FeedPost(tenant).Save(parentPost)
		case err := <-errChan:
			return nil, status.Error(codes.NotFound, "Referenced Post not found. "+err.Error())
		}
	}

	savePostCountPromise := s.db.SocialStats(tenant).UpdatePostCount(userId, 1)

	// wait for async operations to finish.
	<-savePostPromise
	<-saveTagsPromise
	<-savePostCountPromise

	res := &pb.UserPostProto{}
	copier.Copy(res, feedPostModel)

	return res, nil
}

func (s *FeedpostService) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.UserPostProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	postProto := pb.UserPostProto{}

	postChan, errChan := s.db.FeedPost(tenant).FindOneById(req.PostId)
	select {
	case post := <-postChan:
		copier.Copy(&postProto, post)
	case err := <-errChan:
		return nil, status.Error(codes.Internal, err.Error())
	}

	<-extensions.AttachPostUserInfoAsync(s.db, ctx, &postProto, userId, tenant, "default", true)
	return &postProto, nil
}

func (s *FeedpostService) GetFeed(ctx context.Context, req *pb.GetFeedRequest) (*pb.FeedResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	logger.Info("Getting feed for ", zap.String("feedType", req.PostType.String()))

	feed := s.db.FeedPost(tenant).GetFeed(
		req.PostType.String(),
		req.Filters,
		"",
		int64(req.PageNumber),
		int64(req.PageSize))

	feedProto := []*pb.UserPostProto{}
	copier.Copy(&feedProto, feed)

	response := &pb.FeedResponse{Posts: feedProto}

	attachAnswers := (req.PostType == pb.GetFeedRequest_QnA_QUESTION)
	addUserPostActionsPromises := funk.Map(response.Posts, func(x *pb.UserPostProto) chan bool {
		return extensions.AttachPostUserInfoAsync(s.db, ctx, x, userId, tenant, "default", attachAnswers)
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

func (s *FeedpostService) UploadPostMedia(stream pb.UserPost_UploadPostMediaServer) error {
	userId, tenant := auth.GetUserIdAndTenant(stream.Context())
	logger.Info("Uploading post media", zap.String("userId", userId), zap.String("tenant", tenant))
	mediaExtension := "jpg"

	imageData, err := bootUtils.BufferGrpcServerStream(stream, func() ([]byte, error) {
		req, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		mediaExtension = req.MediaExtension
		return req.ChunkData, nil
	})
	if err != nil {
		logger.Error("Failed uploading image", zap.Error(err))
		return err
	}

	// upload imageData to Azure bucket.
	path := fmt.Sprintf("%s/%s/%d-%d.%s", tenant, userId, time.Now().Unix(), rand.Int(), mediaExtension)

	uploadPathChan, errChan := azure.Storage.UploadStream("social-posts", path, imageData)
	select {
	case uploadPath := <-uploadPathChan:
		stream.SendAndClose(&pb.UploadPostMediaResponse{UploadPath: uploadPath})
		return nil
	case err := <-errChan:
		logger.Error("Failed uploading media image.", zap.Error(err))
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *FeedpostService) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	postChan, errChan := s.db.FeedPost(tenant).FindOneById(req.Id)
	select {
	case postEntity := <-postChan:
		if postEntity.UserId != userId {
			return nil, status.Error(codes.PermissionDenied, "User doesn't own the post.")
		}
	case err := <-errChan:
		return nil, status.Error(codes.NotFound, err.Error())
	}

	err := <-s.db.FeedPost(tenant).DeleteById(req.Id)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		return &pb.SocialStatusResponse{
			Status: "success",
		}, nil
	}
}
