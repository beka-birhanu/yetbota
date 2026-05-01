package post

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	ctxYB "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type coordinateDTO struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type orderedPhotoDTO struct {
	ID       string `json:"id"`
	PhotoURL string `json:"photo_url"`
	Position int    `json:"position"`
}

type postDTO struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Likes       int               `json:"likes"`
	Dislikes    int               `json:"dislikes"`
	Comments    int               `json:"comments"`
	UserID      string            `json:"user_id"`
	Tags        []string          `json:"tags"`
	IsQuestion  bool              `json:"is_question"`
	Photos      []orderedPhotoDTO `json:"photos,omitempty"`
	Location    *coordinateDTO    `json:"location,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func setCtxResponse(ctx context.Context, env shared.Envelope) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.Response = env
}

func toPostDTO(p *dbmodels.Post, photos []*postSvc.OrderedPhoto) postDTO {
	if p == nil {
		return postDTO{}
	}

	var loc *coordinateDTO
	if p.Location.Valid && p.Location.Point != nil {
		coords := p.Location.Point.Coords()
		if len(coords) >= 2 {
			// PostGIS point is typically (x=lon, y=lat). We keep names explicit.
			loc = &coordinateDTO{Latitude: coords[1], Longitude: coords[0]}
		}
	}

	var outPhotos []orderedPhotoDTO
	if len(photos) > 0 {
		outPhotos = make([]orderedPhotoDTO, 0, len(photos))
		for _, ph := range photos {
			if ph == nil {
				continue
			}
			outPhotos = append(outPhotos, orderedPhotoDTO{
				ID:       ph.PhotoID,
				PhotoURL: ph.URL,
				Position: ph.Position,
			})
		}
	}

	return postDTO{
		ID:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		Likes:       p.Likes,
		Dislikes:    p.Dislikes,
		Comments:    p.Comments,
		UserID:      p.UserID,
		Tags:        []string(p.Tags),
		IsQuestion:  p.IsQuestion,
		Photos:      outPhotos,
		Location:    loc,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

type postData struct {
	Post postDTO `json:"post"`
}

func encodePostAddHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.AddResponse)
	if !ok || out == nil || out.Post == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: postData{Post: toPostDTO(out.Post, out.Photos)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePostReadHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.ReadResponse)
	if !ok || out == nil || out.Post == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: postData{Post: toPostDTO(out.Post, out.Photos)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

func encodePostUpdateHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.UpdateResponse)
	if !ok || out == nil || out.Post == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: postData{Post: toPostDTO(out.Post, nil)}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}

type voteData struct {
	Likes    int `json:"likes"`
	Dislikes int `json:"dislikes"`
}

func encodePostVoteHTTP(ctx context.Context, w http.ResponseWriter, resp any) error {
	if te, ok := resp.(*toddlerr.Error); ok {
		return te
	}
	out, ok := resp.(*postSvc.PostVoteResponse)
	if !ok || out == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return json.NewEncoder(w).Encode(shared.Envelope{Success: false, Message: "something went wrong"})
	}

	env := shared.Envelope{Success: true, Data: voteData{Likes: out.Likes, Dislikes: out.Dislikes}}
	setCtxResponse(ctx, env)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(env)
}
