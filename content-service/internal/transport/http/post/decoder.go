package post

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

func decodePostAddHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		IsQuestion  bool     `json:"is_question"`
		Photos      []struct {
			PhotoBase64 string `json:"photo_base64"`
			Position    int    `json:"position"`
		} `json:"photos"`
		Location *struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}

	photos := make([]*postSvc.OrderedPhotoUpload, 0, len(in.Photos))
	for _, p := range in.Photos {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p.PhotoBase64))
		if err != nil {
			return nil, badRequest("invalid photo_base64", err)
		}
		photos = append(photos, &postSvc.OrderedPhotoUpload{
			Photo:    b,
			Position: p.Position,
		})
	}

	var lat, lon float64
	if in.Location != nil {
		lat = in.Location.Latitude
		lon = in.Location.Longitude
	}

	req := &postSvc.AddRequest{
		Title:       in.Title,
		Description: in.Description,
		Tags:        in.Tags,
		IsQuestion:  in.IsQuestion,
		Photos:      photos,
		Latitude:    lat,
		Longitude:   lon,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostReadHTTP(ctx context.Context, r *http.Request) (any, error) {
	resolution := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("resolution")))
	if resolution == "" {
		resolution = string(postSvc.PhotoResolutionOriginal)
	}

	req := &postSvc.ReadRequest{
		ID:              r.PathValue("id"),
		PhotoResolution: postSvc.PhotoResolution(resolution),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostUpdateHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Tags         []string `json:"tags"`
		UpsertPhotos []struct {
			PhotoBase64 string `json:"photo_base64"`
			Position    int    `json:"position"`
		} `json:"upsert_photos"`
		Location *struct {
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}

	photos := make([]*postSvc.OrderedPhotoUpload, 0, len(in.UpsertPhotos))
	for _, p := range in.UpsertPhotos {
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(p.PhotoBase64))
		if err != nil {
			return nil, badRequest("invalid photo_base64", err)
		}
		photos = append(photos, &postSvc.OrderedPhotoUpload{
			Photo:    b,
			Position: p.Position,
		})
	}

	var lat, lon float64
	if in.Location != nil {
		lat = in.Location.Latitude
		lon = in.Location.Longitude
	}

	req := &postSvc.UpdateRequest{
		ID:           r.PathValue("id"),
		Title:        in.Title,
		Description:  in.Description,
		Tags:         in.Tags,
		UpsertPhotos: photos,
		Latitude:     lat,
		Longitude:    lon,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodePostVoteHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		VoteType string `json:"vote_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}

	req := &postSvc.PostVoteRequest{
		PostID:   r.PathValue("id"),
		VoteType: strings.ToLower(strings.TrimSpace(in.VoteType)),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func setCtxRequest(ctx context.Context, req any) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.SetRequest(req)
}

func badRequest(publicMsg string, err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     publicMsg,
		ServiceMessage:    err.Error(),
	}
}
