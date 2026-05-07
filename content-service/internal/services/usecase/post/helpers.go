package post

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/geotypes"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"github.com/google/uuid"
	"github.com/twpayne/go-geom"
	"golang.org/x/sync/errgroup"
)

func objectKeyFromURL(keyOrURL string) string {
	if keyOrURL == "" {
		return ""
	}
	if !strings.HasPrefix(keyOrURL, "http://") && !strings.HasPrefix(keyOrURL, "https://") {
		return keyOrURL
	}
	u, err := url.Parse(keyOrURL)
	if err != nil {
		return keyOrURL
	}
	return strings.TrimPrefix(u.Path, "/")
}

func postFromAddReq(req *AddRequest) *dbmodels.Post {
	var location geotypes.NullPoint
	if req.Latitude != 0 || req.Longitude != 0 {
		location = geotypes.NullPoint{Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{req.Longitude, req.Latitude}), Valid: true}
	}

	return &dbmodels.Post{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Description: req.Description,
		Tags:        req.Tags,
		IsQuestion:  req.IsQuestion,
		Location:    location,
		Address:     null.NewString(req.Address, req.Address != ""),
	}
}

func (s *svc) uploadPhotos(ctx context.Context, postID string, photos []*OrderedPhotoUpload) (*uploadPhotosResponse, error) {
	gctx, _ := errgroup.WithContext(ctx)

	res := &uploadPhotosResponse{
		photos:     make(dbmodels.PhotoSlice, len(photos)),
		postPhotos: make(dbmodels.PostPhotoSlice, len(photos)),
	}

	slices.SortFunc(photos, func(a, b *OrderedPhotoUpload) int {
		return a.Position - b.Position
	})

	for i, photo := range photos {
		gctx.Go(func() error {
			if int64(len(photo.Photo)) > constants.MaxUploadSize {
				err := &toddlerr.Error{
					PublicStatusCode:  status.BadRequest,
					ServiceStatusCode: status.BadRequest,
					PublicMessage:     fmt.Sprintf("Image exceeds maximum size of %dMB", constants.MaxUploadSize/constants.MB),
					ServiceMessage:    "image too large",
				}
				return err
			}

			processed, mime, err := utils.ProcessImage(photo.Photo)
			if err != nil {
				return err
			}

			uploadResp, err := s.bucket.UploadFile(ctx, &storage.UploadRequest{
				BucketName:  s.bucketName,
				FileInByte:  processed,
				ContentType: mime,
			})
			if err != nil {
				return err
			}

			id := uuid.NewString()
			res.photos[i] = &dbmodels.Photo{
				ID:             id,
				BucketProvider: dbmodels.PhotoBucketS3,
				MimeType:       uploadResp.ContentType,
				URL:            fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.bucketRegion, uploadResp.FileName),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			res.postPhotos[i] = &dbmodels.PostPhoto{
				PhotoID:  id,
				PostID:   postID,
				Position: photo.Position,
			}

			return nil
		})
	}

	if err := gctx.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *svc) deleteUploads(ctx context.Context, photos dbmodels.PhotoSlice) error {
	gctx, _ := errgroup.WithContext(ctx)

	for _, photo := range photos {
		key := objectKeyFromURL(photo.URL)
		gctx.Go(func() error {
			return s.bucket.RemoveFile(ctx, &storage.DeleteRequest{
				BucketName: s.bucketName,
				FileName:   key,
			})
		})
	}

	if err := gctx.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *svc) assembleOrderedPhoto(ctx context.Context, postPhotos dbmodels.PostPhotoSlice, resolution PhotoResolution) ([]*OrderedPhoto, error) {
	orderedPhotos := make([]*OrderedPhoto, len(postPhotos))
	for i, postPhoto := range postPhotos {
		var photo *dbmodels.Photo
		var err error
		if postPhoto.R != nil && postPhoto.R.Photo != nil {
			photo = postPhoto.R.Photo
		} else {
			photo, err = s.photoRepo.Read(ctx, postPhoto.PhotoID)
			if err != nil {
				return nil, err
			}
		}

		orderedPhotos[i] = &OrderedPhoto{
			ID:       photo.ID,
			PostID:   postPhoto.PostID,
			URL:      pickPhotoURL(photo, resolution),
			Position: postPhoto.Position,
		}
	}

	return orderedPhotos, nil
}

func pickPhotoURL(photo *dbmodels.Photo, res PhotoResolution) string {
	if photo == nil {
		return ""
	}
	switch res {
	case PhotoResolutionMobile:
		if photo.URLMobile.Valid && photo.URLMobile.String != "" {
			return photo.URLMobile.String
		}
		fallthrough
	case PhotoResolutionWeb:
		if photo.URLWeb.Valid && photo.URLWeb.String != "" {
			return photo.URLWeb.String
		}
		fallthrough
	default:
		return photo.URL
	}
}
