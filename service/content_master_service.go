package service

import (
	"context"
	"strings"

	"github.com/Kotlang/socialGo/db"
	pb "github.com/Kotlang/socialGo/generated"
	"github.com/SaiNageswarS/go-api-boot/auth"
	"github.com/SaiNageswarS/go-api-boot/logger"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ContentMasterService struct {
	pb.UnimplementedContentMasterServer
	db *db.SocialDb
}

func NewContentMasterService(socialDB *db.SocialDb) *ContentMasterService {
	return &ContentMasterService{
		db: socialDB,
	}
}

func (s *ContentMasterService) GetContentMaster(ctx context.Context, req *pb.GetContentMasterRequest) (*pb.ContentMasterResponse, error) {
	_, tenant := auth.GetUserIdAndTenant(ctx)

	language := req.Language
	if len(strings.TrimSpace(language)) == 0 {
		language = "english"
	}

	ContentMasterListChan, ContentMasterListErrorChan := s.db.ContentMaster(tenant).FindByLanguage(language)
	list := make([]*pb.ContentMasterProto, 0)

	select {
	case ContentMasterList := <-ContentMasterListChan:
		copier.CopyWithOption(&list, &ContentMasterList, copier.Option{DeepCopy: true})
		return &pb.ContentMasterResponse{
			ContentMasterList: list,
		}, nil
	case err := <-ContentMasterListErrorChan:
		logger.Error("Failed getting content master list", zap.Error(err))
		return nil, status.Error(codes.Internal, "Failed getting content master list")
	}
}
