package feed

import (
	"context"

	pbfeed "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/feed/v1"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoPost(p *dbmodels.Post, photos []*postSvc.OrderedPhoto) *pbfeed.Post {
	if p == nil {
		return nil
	}

	orderedPhotos := make([]*pbfeed.OrderedPhoto, 0, len(photos))
	for _, ph := range photos {
		if ph == nil {
			continue
		}
		orderedPhotos = append(orderedPhotos, &pbfeed.OrderedPhoto{
			Id:       ph.ID,
			Photo:    ph.URL,
			Position: int32(ph.Position),
		})
	}

	post := &pbfeed.Post{
		Id:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Likes:       int32(p.Likes),
		Dislikes:    int32(p.Dislikes),
		Comments:    int32(p.CommentCount),
		UserId:      p.UserID,
		Tags:        []string(p.Tags),
		IsQuestion:  p.IsQuestion,
		Photos:      orderedPhotos,
		Address:     p.Address.String,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}

	if p.Location.Valid && p.Location.Point != nil {
		coords := p.Location.Point.Coords()
		if len(coords) >= 2 {
			post.Location = &pbfeed.Coordinate{
				Latitude:  coords[1],
				Longitude: coords[0],
			}
		}
	}

	return post
}

func deadlineExceeded(ctx context.Context) error {
	if ctx.Err() == context.DeadlineExceeded {
		return grpcstatus.Error(codes.Canceled, "The client canceled the request!")
	}
	return nil
}
