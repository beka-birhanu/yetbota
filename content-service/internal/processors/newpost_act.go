package processors

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"net/http"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type newPostActivity struct {
	postPhotoRepo domainPostphoto.Repository
	photoRepo     domainPhoto.Repository
	bucket        domainStorage.Bucket
	bucketName    string
	bucketRegion  string
}

type newPostActConfig struct {
	PostPhotoRepo domainPostphoto.Repository `validate:"required"`
	PhotoRepo     domainPhoto.Repository     `validate:"required"`
	Bucket        domainStorage.Bucket       `validate:"required"`
	BucketName    string                     `validate:"required"`
	BucketRegion  string                     `validate:"required"`
}

func (c *newPostActConfig) validate() error {
	return validator.Validate.Struct(c)
}

func newNewPostActivity(cfg *newPostActConfig) (*newPostActivity, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &newPostActivity{
		postPhotoRepo: cfg.PostPhotoRepo,
		photoRepo:     cfg.PhotoRepo,
		bucket:        cfg.Bucket,
		bucketName:    cfg.BucketName,
		bucketRegion:  cfg.BucketRegion,
	}, nil
}

// FetchPostPhotoIDs returns IDs of all photos attached to a post, ordered by position.
func (a *newPostActivity) FetchPostPhotoIDs(ctx context.Context, postID string) ([]string, error) {
	postPhotos, err := a.postPhotoRepo.List(ctx,
		&domainPostphoto.Options{PostIDs: []string{postID}},
		&domainPostphoto.SortOptions{
			Field:     domainPostphoto.SortFieldPosition,
			Direction: domainPostphoto.SortDirectionAsc,
		},
	)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(postPhotos))
	for i, pp := range postPhotos {
		ids[i] = pp.PhotoID
	}
	return ids, nil
}

// ProcessPhoto downloads the original, generates mobile/web compressed variants, uploads them,
// and updates the photo DB record. Idempotent: already-processed photos are skipped.
func (a *newPostActivity) ProcessPhoto(ctx context.Context, photoID string) error {
	photo, err := a.photoRepo.Read(ctx, photoID)
	if err != nil {
		return err
	}

	if photo.URLMobile.Valid && photo.URLWeb.Valid {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, photo.URL, nil)
	if err != nil {
		return fmt.Errorf("build request for photo %s: %w", photoID, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("download photo %s: %w", photoID, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("failed to close photo download response body", "photoID", photoID, "error", err)
		}
	}()

	original, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read photo body %s: %w", photoID, err)
	}

	img, _, err := image.Decode(bytes.NewReader(original))
	if err != nil {
		return fmt.Errorf("decode photo %s: %w", photoID, err)
	}

	w := uint(img.Bounds().Dx())
	h := uint(img.Bounds().Dy())
	updated := false

	mobileURL := photo.URLMobile.String
	webURL := photo.URLWeb.String

	if !photo.URLMobile.Valid && (w > constants.MobileImageResolution || h > constants.MobileImageResolution) {
		mobileURL, err = a.compressAndUpload(ctx, original, constants.MobileImageResolution)
		if err != nil {
			return fmt.Errorf("upload mobile variant for photo %s: %w", photoID, err)
		}
		updated = true
	}

	if !photo.URLWeb.Valid && (w > constants.WebImageResolution || h > constants.WebImageResolution) {
		webURL, err = a.compressAndUpload(ctx, original, constants.WebImageResolution)
		if err != nil {
			return fmt.Errorf("upload web variant for photo %s: %w", photoID, err)
		}
		updated = true
	}

	if updated {
		photo.URLMobile = null.StringFrom(mobileURL)
		photo.URLWeb = null.StringFrom(webURL)
		photo.UpdatedAt = time.Now()
		return a.photoRepo.Update(ctx, nil, photo)
	}
	return nil
}

func (a *newPostActivity) compressAndUpload(ctx context.Context, original []byte, maxDim uint) (string, error) {
	compressed, mime, err := utils.CompressToMaxDim(original, maxDim)
	if err != nil {
		return "", err
	}

	uploadResp, err := a.bucket.UploadFile(ctx, &domainStorage.UploadRequest{
		BucketName:  a.bucketName,
		FileInByte:  compressed,
		ContentType: mime,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", a.bucketName, a.bucketRegion, uploadResp.FileName), nil
}
