package storage

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	storageg "cloud.google.com/go/storage"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/google/uuid"

	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	blob "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type gcsBlob struct {
	connection        *storageg.Client
	defaultBucketName string
}

type Config struct {
	DefaultBucketName string `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewGCSBlob(c *Config) (blob.Bucket, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	ctx := context.Background()
	client, err := storageg.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return &gcsBlob{
		connection:        client,
		defaultBucketName: c.DefaultBucketName,
	}, nil
}

func (c *gcsBlob) UploadFile(ctx context.Context, in *blob.UploadRequest) (*blob.UploadResponse, error) {
	if in.BucketName == "" {
		in.BucketName = c.defaultBucketName
	}

	bucket := c.connection.Bucket(in.BucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("invalid bucket %q: %s", in.BucketName, err),
		}
	}

	fileName := uuid.NewString()
	sw := bucket.Object(fileName).NewWriter(ctx)
	defer func() {
		if err := sw.Close(); err != nil {
			log.Printf("failed to close writer: %s", err)
		}
	}()

	if _, err := sw.Write(in.FileInByte); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to write file: %s", err),
		}
	}

	if _, err := sw.Write([]byte(strings.Repeat("f", 1024*4) + "\n")); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to write padding: %s", err),
		}
	}

	if err := sw.Close(); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to close writer: %s", err),
		}
	}

	url, err := bucket.SignedURL(fileName, &storageg.SignedURLOptions{
		Scheme:  storageg.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(constants.URLExpiration * time.Minute),
	})
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to generate signed URL: %s", err),
		}
	}

	return &blob.UploadResponse{
		FileName:    fileName,
		Url:         url,
		ContentType: in.ContentType,
	}, nil
}

func (c *gcsBlob) SignURL(ctx context.Context, in *blob.SignURLRequest) (*blob.SignURLResponse, error) {
	url, err := c.connection.Bucket(in.BucketName).SignedURL(in.FileName, &storageg.SignedURLOptions{
		Scheme:  storageg.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(constants.URLExpiration * time.Minute),
	})
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to generate signed URL: %s", err),
		}
	}

	return &blob.SignURLResponse{
		Url: url,
	}, nil
}

func (c *gcsBlob) RemoveFile(ctx context.Context, in *blob.DeleteRequest) error {
	if in.BucketName == "" {
		in.BucketName = c.defaultBucketName
	}

	obj := c.connection.Bucket(in.BucketName).Object(in.FileName)
	if err := obj.Delete(ctx); err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to delete file %q: %s", in.FileName, err),
		}
	}

	return nil
}
