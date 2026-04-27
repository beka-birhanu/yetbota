package post

import (
	"context"

	pb "github.com/beka-birhanu/yetbota/common/proto/generated/go/content/v1"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

func decodeAddReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.AddRequest)

	photos := make([]*postSvc.OrderedPhotoUpload, 0, len(in.GetPhotos()))
	for _, p := range in.GetPhotos() {
		photos = append(photos, &postSvc.OrderedPhotoUpload{
			Photo:    p.GetPhoto(),
			Position: int(p.GetPosition()),
		})
	}

	var lat, lon float64
	if loc := in.GetLocation(); loc != nil {
		lat = loc.GetLatitude()
		lon = loc.GetLongitude()
	}

	return &postSvc.AddRequest{
		Title:       in.GetTitle(),
		Description: in.GetDescription(),
		Tags:        in.GetTags(),
		IsQuestion:  in.GetIsQuestion(),
		Photos:      photos,
		Latitude:    lat,
		Longitude:   lon,
	}, nil
}

func decodeReadReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.ReadRequest)
	return &postSvc.ReadRequest{
		ID:              in.GetId(),
		PhotoResolution: mapPhotoResolutionFromProto(in.GetResolution()),
	}, nil
}

func decodeUpdateReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.UpdateRequest)

	photos := make([]*postSvc.OrderedPhotoUpload, 0, len(in.GetUpsertPhotos()))
	for _, p := range in.GetUpsertPhotos() {
		photos = append(photos, &postSvc.OrderedPhotoUpload{
			Photo:    p.GetPhoto(),
			Position: int(p.GetPosition()),
		})
	}

	var lat, lon float64
	if loc := in.GetLocation(); loc != nil {
		lat = loc.GetLatitude()
		lon = loc.GetLongitude()
	}

	return &postSvc.UpdateRequest{
		ID:           in.GetId(),
		Title:        in.GetTitle(),
		Description:  in.GetDescription(),
		Tags:         in.GetTags(),
		UpsertPhotos: photos,
		Latitude:     lat,
		Longitude:    lon,
	}, nil
}

func decodeVoteReq(_ context.Context, req any) (any, error) {
	in := req.(*pb.VotePostRequest)
	return &postSvc.PostVoteRequest{
		PostID:   in.GetPostId(),
		VoteType: mapPostVoteTypeFromProto(in.GetVoteType()),
	}, nil
}
