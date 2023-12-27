package service

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/extensions"
	notificationPb "github.com/Kotlang/socialGo/generated/notification"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/Kotlang/socialGo/models"
	s3client "github.com/Kotlang/socialGo/s3Client"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/azure"
	"github.com/SaiNageswarS/go-api-boot/bootUtils"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FeedpostService struct {
	socialPb.UnimplementedUserPostServer
	db db.SocialDbInterface
}

func NewFeedpostService(db db.SocialDbInterface) *FeedpostService {
	return &FeedpostService{
		db: db,
	}
}

func (s *FeedpostService) CreatePost(ctx context.Context, req *socialPb.UserPostRequest) (*socialPb.UserPostProto, error) {
	err := ValidateUserPostRequest(req)
	if err != nil {
		return nil, err
	}

	userId, tenant := auth.GetUserIdAndTenant(ctx)

	// map proto to model.
	feedPostModel := s.db.FeedPost(tenant).GetModel(req)
	feedPostModel.UserId = userId
	feedPostModel.PostType = req.PostType.String()

	if len(strings.TrimSpace(feedPostModel.Language)) == 0 {
		feedPostModel.Language = "english"
	}

	// save post.
	savePostPromise := s.db.FeedPost(tenant).Save(feedPostModel)

	// save tags.
	saveTagsPromise := extensions.SaveTags(s.db, tenant, req.Tags)

	savePostCountPromise := s.db.SocialStats(tenant).UpdatePostCount(userId, 1)

	// wait for async operations to finish.
	if err := <-savePostPromise; err != nil {
		return nil, err
	}
	<-saveTagsPromise
	if err := <-savePostCountPromise; err != nil {
		return nil, err
	}

	feedPostModelChan, errChan := s.db.FeedPost(tenant).FindOneById(feedPostModel.PostId)

	select {
	case feedPostModel := <-feedPostModelChan:
		res := &socialPb.UserPostProto{}
		copier.Copy(res, feedPostModel)

		attachAuthorInfoPromise := extensions.AttachPostUserInfoAsync(s.db, ctx, res, userId, tenant)
		<-attachAuthorInfoPromise

		authorName := ""
		if res.AuthorInfo != nil {
			authorName = res.AuthorInfo.Name
		}
		tag := ""
		if len(res.Tags) > 0 {
			tag = res.Tags[0]
		}
		imageURL := ""
		if len(res.MediaUrls) > 0 {
			imageURL = res.MediaUrls[0].Url
		}

		err := <-extensions.RegisterEvent(ctx, &notificationPb.RegisterEventRequest{
			EventType: "post.created",
			Title:     fmt.Sprintf("%s ने %s पर अपना %s शेयर किया है", authorName, tag, feedPostModel.PostType),
			Body:      fmt.Sprintf("%s आपके विचार सुनना चाहता है! एप खोलें और कमेंट करें!", authorName),
			ImageURL:  imageURL,
			TemplateParameters: map[string]string{
				"postId": feedPostModel.PostId,
			},
			Topic:       fmt.Sprintf("%s.post.created", tenant),
			TargetUsers: []string{userId},
		})
		if err != nil {
			logger.Error("Failed to register event", zap.Error(err))
		}

		return res, nil
	case err := <-errChan:
		return nil, status.Error(codes.Internal, err.Error())
	}
}

func (s *FeedpostService) GetPost(ctx context.Context, req *socialPb.GetPostRequest) (*socialPb.UserPostProto, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	postProto := socialPb.UserPostProto{}

	filters := bson.M{}
	filters["_id"] = req.PostId
	filters["isDeleted"] = false

	postChan, errChan := s.db.FeedPost(tenant).FindOne(filters)

	select {
	case post := <-postChan:
		copier.Copy(&postProto, post)
	case err := <-errChan:
		return nil, status.Error(codes.Internal, err.Error())
	}

	<-extensions.AttachPostUserInfoAsync(s.db, ctx, &postProto, userId, tenant)
	return &postProto, nil
}

func (s *FeedpostService) GetFeed(ctx context.Context, req *socialPb.GetFeedRequest) (*socialPb.FeedResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	if req.PageSize == 0 {
		req.PageSize = 10
	}
	if req.Filters != nil {
		logger.Info("Getting feed for ", zap.String("feedType", req.Filters.PostType.String()))
	} else {
		err := "PostType filters is not provided"
		logger.Error(err)
		return nil, status.Error(codes.InvalidArgument, err)
	}
	req.Filters.UserId = userId
	feed := s.db.FeedPost(tenant).GetFeed(
		req.Filters,
		int64(req.PageNumber),
		int64(req.PageSize))

	feedProto := []*socialPb.UserPostProto{}
	copier.Copy(&feedProto, feed)

	response := &socialPb.FeedResponse{Posts: feedProto}

	attachPostInfoPromise := extensions.AttachMultiplePostUserInfoAsync(s.db, ctx, response.Posts, userId, tenant)
	<-attachPostInfoPromise
	response.PageNumber = req.PageNumber
	response.PageSize = req.PageSize
	return response, nil
}

func (s *FeedpostService) DeletePost(ctx context.Context, req *socialPb.DeletePostRequest) (*socialPb.SocialStatusResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)

	postChan, errChan := s.db.FeedPost(tenant).FindOneById(req.Id)
	postEntity := &models.FeedPostModel{}

	IsUserAdmin := <-extensions.IsUserAdmin(ctx, userId)
	select {
	case postEntity = <-postChan:
		if postEntity.UserId != userId && !IsUserAdmin {
			return nil, status.Error(codes.PermissionDenied, "User doesn't own the post.")
		}
	case err := <-errChan:
		return nil, status.Error(codes.NotFound, err.Error())
	}

	postEntity.IsDeleted = true

	err := <-s.db.FeedPost(tenant).Save(postEntity)

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	} else {
		return &socialPb.SocialStatusResponse{
			Status: "success",
		}, nil
	}
}

func (s *FeedpostService) GetTags(ctx context.Context, req *socialPb.GetTagsRequest) (*socialPb.TagListResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	tags := s.db.Tag(tenant).FindTagsRanked()

	res := &socialPb.TagListResponse{}
	copier.Copy(&res.Tags, tags)

	return res, nil
}

func (s *FeedpostService) GetMediaUploadUrl(ctx context.Context, req *socialPb.MediaUploadRequest) (*socialPb.MediaUploadURL, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	uploadUrl, downloadUrl := s3client.GetPresignedUrlForPosts(tenant, userId, req.MediaExtension)
	return &socialPb.MediaUploadURL{
		UploadUrl: uploadUrl,
		MediaUrl:  downloadUrl,
	}, nil
}

func (s *FeedpostService) UploadPostMedia(stream socialPb.UserPost_UploadPostMediaServer) error {
	userId, tenant := auth.GetUserIdAndTenant(stream.Context())
	logger.Info("Uploading post media", zap.String("userId", userId), zap.String("tenant", tenant))
	maxFileSize := 50 * 1024 * 1024

	allowedMimeTypes := []string{"image/jpeg", "image/png", "video/avi", "video/mp4", "video/webm"}
	imageData, contentType, err := bootUtils.BufferGrpcServerStream(
		allowedMimeTypes,
		maxFileSize,
		func() ([]byte, error) {
			err := bootUtils.StreamContextError(stream.Context())
			if err != nil {
				return nil, err
			}

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

	mediaExtension := bootUtils.GetFileExtension(contentType)
	// upload imageData to Azure bucket.
	path := fmt.Sprintf("%s/%s/%d-%d.%s", tenant, userId, time.Now().Unix(), rand.Int(), mediaExtension)

	uploadPathChan, errChan := azure.Storage.UploadStream("social-posts", path, imageData)
	select {
	case uploadPath := <-uploadPathChan:
		stream.SendAndClose(&socialPb.UploadPostMediaResponse{UploadPath: uploadPath})
		return nil
	case err := <-errChan:
		logger.Error("Failed uploading media image.", zap.Error(err))
		return status.Error(codes.Internal, err.Error())
	}
}

func (s *FeedpostService) ParsePost(ctx context.Context, req *socialPb.UserPostRequest) (*socialPb.UserPostRequest, error) {
	links := <-extensions.GetLinks(req.Post)

	links = append(links, funk.Map(req.MediaUrls, func(x *socialPb.MediaUrl) string { return x.Url }).([]string)...)
	links = funk.UniqString(links)
	mediaUrlsChan, webPreviewsChan := extensions.GeneratePreviews(links)

	mediaUrls := <-mediaUrlsChan
	newUserPost := &socialPb.UserPostRequest{}
	copier.CopyWithOption(newUserPost, req, copier.Option{
		DeepCopy: true,
	})
	newUserPost.MediaUrls = mediaUrls
	newUserPost.WebPreviews = <-webPreviewsChan
	return newUserPost, nil
}
