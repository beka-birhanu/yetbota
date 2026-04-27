package post

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
)

type OrderedPhotoUpload struct {
	Photo    []byte
	Position int
}

type OrderedPhoto struct {
	PhotoID  string
	URL      string
	Position int
}

type PhotoResolution string

const (
	PhotoResolutionUnspecified PhotoResolution = ""
	PhotoResolutionOriginal    PhotoResolution = "ORIGINAL"
	PhotoResolutionMobile      PhotoResolution = "MOBILE"
	PhotoResolutionWeb         PhotoResolution = "WEB"
)

// Add

type AddRequest struct {
	Title       string   `validate:"required"`
	Description string   `validate:"required"`
	Tags        []string `validate:"dive,min=1,max=20"`
	IsQuestion  bool
	Photos      []*OrderedPhotoUpload
	Latitude    float64
	Longitude   float64
}

func (r *AddRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AddResponse struct {
	Post   *dbmodels.Post
	Photos []*OrderedPhoto
}

// Read

type ReadRequest struct {
	ID              string          `validate:"required"`
	PhotoResolution PhotoResolution `validate:"required,oneof=ORIGINAL MOBILE WEB"`
}

func (r *ReadRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ReadResponse struct {
	Post   *dbmodels.Post
	Photos []*OrderedPhoto
}

// Update

type UpdateRequest struct {
	ID           string   `validate:"required"`
	Title        string   `validate:"required"`
	Description  string   `validate:"required"`
	Tags         []string `validate:"dive,min=1,max=20"`
	UpsertPhotos []*OrderedPhotoUpload
	Latitude     float64 `validate:"omitempty,min=-90,max=90"`
	Longitude    float64 `validate:"omitempty,min=-180,max=180"`
}

func (r *UpdateRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UpdateResponse struct {
	Post *dbmodels.Post
}

type uploadPhotosResponse struct {
	photos     dbmodels.PhotoSlice
	postPhotos dbmodels.PostPhotoSlice
}

type PostVoteRequest struct {
	PostID   string `validate:"required"`
	VoteType string `validate:"required" oneof:"like dislike"`
}

func (r *PostVoteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}

	return nil
}

type PostVoteResponse struct {
	Likes    int
	Dislikes int
}
