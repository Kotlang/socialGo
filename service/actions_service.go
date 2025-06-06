package service

import (
	"context"

	"github.com/Kotlang/socialGo/db"
	"github.com/Kotlang/socialGo/extensions"
	socialPb "github.com/Kotlang/socialGo/generated/social"
	"github.com/Kotlang/socialGo/models"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"github.com/thoas/go-funk"
	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ActionsService struct {
	socialPb.UnimplementedActionsServer
	db db.SocialDbInterface
}

func NewActionsService(db db.SocialDbInterface) *ActionsService {
	return &ActionsService{
		db: db,
	}
}

func (s *ActionsService) React(ctx context.Context, req *socialPb.ReactRequest) (*socialPb.SocialStatusResponse, error) {
	if req.Reaction == "" || req.EntityId == "" {
		return nil, status.Error(codes.InvalidArgument, "Fields missing")
	}

	userId, tenant := auth.GetUserIdAndTenant(ctx)
	reactionModel := getExistingReaction(s.db, tenant, models.GetReactionId(userId, req.EntityId))

	//check if reaction already exists in db
	reactionExists := stringInArray(req.Reaction, reactionModel.Reaction)
	if reactionExists {
		return nil, status.Error(codes.AlreadyExists, "Reaction already exists")
	}

	isNewReaction := false
	if len(reactionModel.Id()) == 0 || !reactionExists {
		isNewReaction = true
	}

	//merge old values into new values
	reactionModel.Reaction = append(reactionModel.Reaction, req.Reaction)
	reactionModel.EntityId = req.EntityId
	reactionModel.ReactionOn = req.ReactionOn.String()
	reactionModel.UserId = userId

	// increment numReacts of the entity
	if isNewReaction {
		switch req.ReactionOn {
		case socialPb.EntityTypes_POST:
			feedPostChan, errChan := s.db.FeedPost(tenant).FindOneById(req.EntityId)
			select {
			case feedPost := <-feedPostChan:
				if feedPost.NumReacts == nil {
					feedPost.NumReacts = make(map[string]int64)
				}
				feedPost.NumReacts[req.Reaction] = feedPost.NumReacts[req.Reaction] + 1
				<-s.db.FeedPost(tenant).Save(feedPost)
			case err := <-errChan:
				logger.Error("Probably post not found", zap.Error(err))
				return nil, err
			}
		case socialPb.EntityTypes_EVENT:
			eventChan, errChan := s.db.Event(tenant).FindOneById(req.EntityId)
			select {
			case event := <-eventChan:
				if event.NumReacts == nil {
					event.NumReacts = make(map[string]int64)
				}
				event.NumReacts[req.Reaction] = event.NumReacts[req.Reaction] + 1
				<-s.db.Event(tenant).Save(event)
			case err := <-errChan:
				logger.Error("Probably event not found", zap.Error(err))
				return nil, err
			}
		case socialPb.EntityTypes_COMMENT:
			commentChan, errChan := s.db.Comment(tenant).FindOneById(req.EntityId)
			select {
			case comment := <-commentChan:
				if comment.NumReacts == nil {
					comment.NumReacts = make(map[string]int64)
				}
				comment.NumReacts[req.Reaction] = comment.NumReacts[req.Reaction] + 1
				<-s.db.Comment(tenant).Save(comment)
			case err := <-errChan:
				logger.Error("Probably comment not found", zap.Error(err))
				return nil, err
			}
		}
	}

	errChan := s.db.React(tenant).Save(reactionModel)
	UpdateReactPromise := s.db.SocialStats(tenant).UpdateReactCount(userId, 1)
	<-UpdateReactPromise
	err := <-errChan
	if err != nil {
		logger.Error("Failed saving reaction", zap.Error(err))
		return nil, err
	}
	return &socialPb.SocialStatusResponse{Status: "success"}, nil
}

func (s *ActionsService) UnReact(ctx context.Context, req *socialPb.ReactRequest) (*socialPb.SocialStatusResponse, error) {
	if req.Reaction == "" || req.EntityId == "" {
		return nil, status.Error(codes.InvalidArgument, "Fields missing")
	}

	userId, tenant := auth.GetUserIdAndTenant(ctx)

	reactionResChan, errResChan := s.db.React(tenant).FindOneById(models.GetReactionId(userId, req.EntityId))
	reactionModel := &models.ReactionModel{}

	// check if reaction exists in db, if yes, remove it
	select {
	case reactionModel = <-reactionResChan:
		newReactionArray := removeElement(reactionModel.Reaction, req.Reaction)
		if len(reactionModel.Reaction) == len(newReactionArray) {
			return nil, status.Error(codes.NotFound, "Reaction not found")
		}

		// if reaction array is empty, delete the reaction
		var err error
		if len(newReactionArray) == 0 {
			err = <-s.db.React(tenant).DeleteById(reactionModel.Id())
		} else {
			reactionModel.Reaction = newReactionArray
			err = <-s.db.React(tenant).Save(reactionModel)
		}
		if err != nil {
			logger.Error("Error deleting reaction", zap.Error(err))
			return &socialPb.SocialStatusResponse{Status: "fail"}, err
		}

	case err := <-errResChan:
		logger.Error("Probably Reaction not found", zap.Error(err))
		return nil, err
	}

	// decrement numReacts of the entity
	switch reactionModel.ReactionOn {
	case socialPb.EntityTypes_POST.String():
		feedPostChan, errChan := s.db.FeedPost(tenant).FindOneById(reactionModel.EntityId)
		select {
		case feedPost := <-feedPostChan:
			feedPost.NumReacts[req.Reaction] = feedPost.NumReacts[req.Reaction] - 1
			<-s.db.FeedPost(tenant).Save(feedPost)
		case err := <-errChan:
			logger.Error("Probably post not found", zap.Error(err))
			return nil, err
		}
	case socialPb.EntityTypes_EVENT.String():
		eventChan, errChan := s.db.Event(tenant).FindOneById(req.EntityId)
		select {
		case event := <-eventChan:
			event.NumReacts[req.Reaction] = event.NumReacts[req.Reaction] - 1
			<-s.db.Event(tenant).Save(event)
		case err := <-errChan:
			logger.Error("Probably event not found", zap.Error(err))
			return nil, err
		}
	case socialPb.EntityTypes_COMMENT.String():
		commentChan, errChan := s.db.Comment(tenant).FindOneById(req.EntityId)
		select {
		case comment := <-commentChan:
			comment.NumReacts[req.Reaction] = comment.NumReacts[req.Reaction] - 1
			<-s.db.Comment(tenant).Save(comment)
		case err := <-errChan:
			logger.Error("Probably comment not found", zap.Error(err))
			return nil, err
		}
	}
	UpdateReactPromise := s.db.SocialStats(tenant).UpdateReactCount(userId, -1)
	<-UpdateReactPromise
	return &socialPb.SocialStatusResponse{Status: "success"}, nil
}

func (s *ActionsService) Comment(ctx context.Context, req *socialPb.CommentRequest) (*socialPb.CommentProto, error) {
	if req.Content == "" || req.ParentId == "" {
		return nil, status.Error(codes.InvalidArgument, "Fields missing")
	}

	userId, tenant := auth.GetUserIdAndTenant(ctx)
	commentModel := s.db.Comment(tenant).GetModel(req)

	commentModel.UserId = userId
	commentModel.CommentOn = req.CommentOn.String()

	//update the numReplies of the parent
	switch req.CommentOn {
	case socialPb.EntityTypes_POST:
		feedPostChan, errChan := s.db.FeedPost(tenant).FindOneById(req.ParentId)
		select {
		case feedPost := <-feedPostChan:
			feedPost.NumReplies = feedPost.NumReplies + 1
			<-s.db.FeedPost(tenant).Save(feedPost)
		case err := <-errChan:
			logger.Error("Probably post not found", zap.Error(err))
			return nil, err
		}
	case socialPb.EntityTypes_EVENT:
		eventChan, errChan := s.db.Event(tenant).FindOneById(req.ParentId)
		select {
		case event := <-eventChan:
			event.NumReplies = event.NumReplies + 1
			<-s.db.Event(tenant).Save(event)
		case err := <-errChan:
			logger.Error("Probably event not found", zap.Error(err))
			return nil, err
		}
	case socialPb.EntityTypes_COMMENT:
		commentChan, errChan := s.db.Comment(tenant).FindOneById(req.ParentId)
		select {
		case comment := <-commentChan:
			comment.NumReplies = comment.NumReplies + 1
			<-s.db.Comment(tenant).Save(comment)
		case err := <-errChan:
			logger.Error("Probably comment not found", zap.Error(err))
			return nil, err
		}
	}

	commentAsyncSaveRequest := s.db.Comment(tenant).Save(commentModel)
	<-commentAsyncSaveRequest

	updateCommentPromise := s.db.SocialStats(tenant).UpdateCommentsCount(userId, 1)
	<-updateCommentPromise

	// fetch the saved comment
	commentProto := &socialPb.CommentProto{}
	commentResChan, errChan := s.db.Comment(tenant).FindOneById(commentModel.Id())
	select {
	case comment := <-commentResChan:
		copier.Copy(commentProto, comment)
	case err := <-errChan:
		logger.Error("Probably comment not found", zap.Error(err))
		return nil, err
	}
	<-extensions.AttachCommentUserInfoAsync(s.db, ctx, commentProto, userId, tenant, "default")
	return commentProto, nil
}

// TODO: delete nested comments, write extension for delete
func (s *ActionsService) DeleteComment(ctx context.Context, req *socialPb.IdRequest) (*socialPb.SocialStatusResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	commentResChan, errResChan := s.db.Comment(tenant).FindOneById(req.Id)
	comment := &models.CommentModel{}

	//mark comment as deleted
	select {
	case comment = <-commentResChan:
		comment.IsDeleted = true
		<-s.db.Comment(tenant).Save(comment)

		updateCommentPromise := s.db.SocialStats(tenant).UpdateCommentsCount(comment.UserId, -1)
		<-updateCommentPromise
	case err := <-errResChan:
		logger.Error("Probably comment not found", zap.Error(err))
		return nil, err
	}

	//reduce numReplies of parent
	switch comment.CommentOn {
	case socialPb.EntityTypes_POST.String():
		feedPostChan, errChan := s.db.FeedPost(tenant).FindOneById(comment.ParentId)
		select {
		case feedPost := <-feedPostChan:
			feedPost.NumReplies = feedPost.NumReplies - 1
			<-s.db.FeedPost(tenant).Save(feedPost)
		case err := <-errChan:
			logger.Error("Probably post not found", zap.Error(err))
			return nil, err
		}
	case socialPb.EntityTypes_EVENT.String():
		eventChan, errChan := s.db.Event(tenant).FindOneById(comment.ParentId)
		select {
		case event := <-eventChan:
			event.NumReplies = event.NumReplies - 1
			<-s.db.Event(tenant).Save(event)
		case err := <-errChan:
			logger.Error("Probably event not found", zap.Error(err))
			return nil, err
		}
	case socialPb.EntityTypes_COMMENT.String():
		commentChan, errChan := s.db.Comment(tenant).FindOneById(comment.ParentId)
		select {
		case comment := <-commentChan:
			comment.NumReplies = comment.NumReplies - 1
			<-s.db.Comment(tenant).Save(comment)
		case err := <-errChan:
			logger.Error("Probably comment not found", zap.Error(err))
			return nil, err
		}
	}

	return &socialPb.SocialStatusResponse{Status: "success"}, nil
}

// TODO: fetch nested comments, write extension for fetch
func (s *ActionsService) FetchComments(ctx context.Context, req *socialPb.CommentFetchRequest) (*socialPb.CommentsFetchResponse, error) {
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	comments := s.db.Comment(tenant).GetComments(req.ParentId, req.UserId, int64(req.PageNumber), int64(req.PageSize))
	commentProtos := []*socialPb.CommentProto{}
	copier.Copy(&commentProtos, &comments)

	response := &socialPb.CommentsFetchResponse{Comments: commentProtos, PageNumber: req.PageNumber, PageSize: req.PageSize}

	addUserInfoActionsPromises := funk.Map(response.Comments, func(x *socialPb.CommentProto) chan bool {
		return extensions.AttachCommentUserInfoAsync(s.db, ctx, x, userId, tenant, "default")
	}).([]chan bool)
	for _, promise := range addUserInfoActionsPromises {
		<-promise
	}
	return response, nil
}

func (s *ActionsService) FetchCommentById(ctx context.Context, req *socialPb.IdRequest) (*socialPb.CommentProto, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "Field Id missing")
	}
	userId, tenant := auth.GetUserIdAndTenant(ctx)
	filter := bson.M{
		"_id":       req.Id,
		"isDeleted": false,
	}
	commentResChan, errChan := s.db.Comment(tenant).FindOne(filter)
	select {
	case comment := <-commentResChan:
		commentProto := &socialPb.CommentProto{}
		copier.Copy(commentProto, comment)
		<-extensions.AttachCommentUserInfoAsync(s.db, ctx, commentProto, userId, tenant, "default")
		return commentProto, nil
	case err := <-errChan:
		logger.Error("Probably comment not found", zap.Error(err))
		return nil, err
	}
}

func getExistingReaction(db db.SocialDbInterface, tenant string, reactionId string) *models.ReactionModel {
	reaction := &models.ReactionModel{}

	reactionResChan, errChan := db.React(tenant).FindOneById(reactionId)

	select {
	case reactionRes := <-reactionResChan:
		reaction = reactionRes
	case err := <-errChan:
		logger.Error("Reaction not found", zap.Error(err))
	}
	return reaction
}

func removeElement(array []string, element string) []string {
	// Create a new array to store the filtered elements.
	newArray := []string{}

	// Iterate over the original array and add all elements except the one to be removed to the new array.
	for _, v := range array {
		if v != element {
			newArray = append(newArray, v)
		}
	}

	// Return the new array.
	return newArray
}

func stringInArray(element string, array []string) bool {
	for _, item := range array {
		if element == item {
			return true
		}
	}
	return false
}
