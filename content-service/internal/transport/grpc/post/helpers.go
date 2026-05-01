package post

import (
	"context"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/v1"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func deadlineExceeded(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return status.Error(codes.Canceled, "The client canceled the request!")
	}
	return nil
}

func postToProto(p *dbmodels.Post, photos []*postSvc.OrderedPhoto) *pb.Post {
	orderedPhotos := make([]*pb.OrderedPhoto, 0, len(photos))
	for _, photo := range photos {
		orderedPhotos = append(orderedPhotos, &pb.OrderedPhoto{
			Photo:    photo.URL,
			Position: int32(photo.Position),
		})
	}

	var loc *pb.Coordinate
	if p.Location.Valid && p.Location.Point != nil {
		coords := p.Location.Point.FlatCoords()
		loc = &pb.Coordinate{
			Latitude:  coords[0],
			Longitude: coords[1],
		}
	}

	return &pb.Post{
		Id:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Likes:       int32(p.Likes),
		Dislikes:    int32(p.Dislikes),
		Comments:    int32(p.Comments),
		UserId:      p.UserID,
		Tags:        p.Tags,
		IsQuestion:  p.IsQuestion,
		Photos:      orderedPhotos,
		Location:    loc,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
		Address:     p.Address.String,
	}
}

func mapPhotoResolutionFromProto(r pb.PhotoResolution) postSvc.PhotoResolution {
	switch r {
	case pb.PhotoResolution_PHOTO_RESOLUTION_ORIGINAL:
		return postSvc.PhotoResolutionOriginal
	case pb.PhotoResolution_PHOTO_RESOLUTION_MOBILE:
		return postSvc.PhotoResolutionMobile
	case pb.PhotoResolution_PHOTO_RESOLUTION_WEB:
		return postSvc.PhotoResolutionWeb
	default:
		return postSvc.PhotoResolutionUnspecified
	}
}

func mapPostVoteTypeFromProto(v pb.PostVoteType) string {
	switch v {
	case pb.PostVoteType_POST_VOTE_TYPE_LIKE:
		return dbmodels.PostVoteTypeLike
	case pb.PostVoteType_POST_VOTE_TYPE_DISLIKE:
		return dbmodels.PostVoteTypeDislike
	default:
		return dbmodels.PostVoteTypeLike
	}
}

func mapSortFieldFromProto(f pb.PostSortField) domainPost.ListSortField {
	switch f {
	case pb.PostSortField_POST_SORT_FIELD_LIKES:
		return domainPost.ListSortFieldLikes
	case pb.PostSortField_POST_SORT_FIELD_DISLIKES:
		return domainPost.ListSortFieldDislikes
	case pb.PostSortField_POST_SORT_FIELD_COMMENTS:
		return domainPost.ListSortFieldComments
	default:
		return domainPost.ListSortFieldCreatedAt
	}
}

func mapSortDirFromProto(d pb.SortDirection) domainPost.ListSortDir {
	if d == pb.SortDirection_SORT_DIRECTION_ASC {
		return domainPost.ListSortDirAsc
	}
	return domainPost.ListSortDirDesc
}
