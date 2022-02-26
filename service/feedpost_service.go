package service

import (
	"context"
	"fmt"
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
	feedPostModel := s.db.FeedPost(tenant).GetModel(req).(*models.FeedPostModel)
	feedPostModel.UserId = userId
	feedPostModel.PostType = pb.UserPostRequest_PostType_name[int32(req.PostType)]

	// save post.
	savePostPromise := s.db.FeedPost(tenant).Save(feedPostModel)

	// save tags.
	saveTagsPromise := extensions.SaveTags(s.db, tenant, req.Tags)

	// if it is a comment/answer increment numReplies
	if len(feedPostModel.ReferencePost) > 0 {
		parentPostRes := <-s.db.FeedPost(tenant).FindOneById(feedPostModel.ReferencePost)
		parentPost := parentPostRes.Value.(*models.FeedPostModel)
		parentPost.NumReplies = parentPost.NumReplies + 1
		<-s.db.FeedPost(tenant).Save(parentPost)
	}

	// wait for async operations to finish.
	<-savePostPromise
	<-saveTagsPromise

	res := &pb.UserPostProto{}
	copier.Copy(res, feedPostModel)

	return res, nil
}

func (s *FeedpostService) GetPost(ctx context.Context, req *pb.GetPostRequest) (*pb.UserPostProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	post := <-s.db.FeedPost(tenant).FindOneById(req.PostId)

	if post.Err != nil {
		return nil, status.Error(codes.Internal, post.Err.Error())
	}

	postProto := pb.UserPostProto{}
	copier.Copy(&postProto, post.Value)

	authClient := extensions.NewAuthClient(ctx)

	<-extensions.AttachPostUserInfoAsync(s.db, authClient, &postProto, userId, tenant, "default", true)
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
	authClient := extensions.NewAuthClient(ctx)

	attachAnswers := (req.PostType == pb.GetFeedRequest_QnA_QUESTION)
	addUserPostActionsPromises := funk.Map(response.Posts, func(x *pb.UserPostProto) chan bool {
		return extensions.AttachPostUserInfoAsync(s.db, authClient, x, userId, tenant, "default", attachAnswers)
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
	imageData, err := bootUtils.BufferGrpcServerStream(stream, func() ([]byte, error) {
		req, err := stream.Recv()
		if err != nil {
			return nil, err
		}
		return req.ChunkData, nil
	})
	if err != nil {
		logger.Error("Failed uploading image", zap.Error(err))
		return err
	}

	// upload imageData to Azure bucket.
	path := fmt.Sprintf("%s/%s/%d.jpg", tenant, userId, time.Now().Unix())
	res := <-azure.Storage.UploadStream("social-posts", path, imageData)

	if res.Err != nil {
		logger.Error("Failed uploading media image.", zap.Error(res.Err))
		return res.Err
	}
	stream.SendAndClose(&pb.UploadPostMediaResponse{UploadPath: res.Value.(string)})
	return nil
}

func (s *FeedpostService) DeletePost(ctx context.Context, req *pb.DeletePostRequest) (*pb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	post := <-s.db.FeedPost(tenant).FindOneById(req.Id)
	if post.Err != nil {
		return nil, status.Error(codes.NotFound, post.Err.Error())
	}

	postEntity := post.Value.(*models.FeedPostModel)
	if postEntity.UserId != userId {
		return nil, status.Error(codes.PermissionDenied, "User doesn't own the post.")
	}

	err := <-s.db.FeedPost(tenant).DeleteById(req.Id)

	if err != nil {
		return nil, err
	} else {
		return &pb.SocialStatusResponse{
			Status: "success",
		}, nil
	}
}
