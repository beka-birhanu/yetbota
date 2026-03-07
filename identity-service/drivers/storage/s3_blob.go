package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	tmtypes "github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	blob "github.com/beka-birhanu/yetbota/identity-service/internal/domain/storage"
	"github.com/google/uuid"
)

type s3Blob struct {
	client          *s3.Client
	transferManager *transfermanager.Client
	BucketNameImage string
}

type S3Config struct {
	DefaultBucketName string     `validate:"required"`
	AwsConfig         aws.Config `validate:"required"`
}

func (c *S3Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewS3Blob(c *S3Config) (blob.Bucket, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(c.AwsConfig)
	return &s3Blob{
		client:          client,
		transferManager: transfermanager.New(client),
		BucketNameImage: c.DefaultBucketName,
	}, nil
}

func (c *s3Blob) UploadFile(ctx context.Context, in *blob.UploadRequest) (*blob.UploadResponse, error) {
	if in.BucketName == "" {
		in.BucketName = c.BucketNameImage
	}

	fileName := uuid.NewString()
	_, err := c.transferManager.UploadObject(ctx, &transfermanager.UploadObjectInput{
		Bucket:            aws.String(in.BucketName),
		Key:               aws.String(fileName),
		Body:              strings.NewReader(string(in.FileInByte)),
		ACL:               tmtypes.ObjectCannedACLPrivate,
		ChecksumAlgorithm: tmtypes.ChecksumAlgorithmSha256,
		ContentType:       aws.String(in.ContentType),
	})
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to upload file: %s", err),
		}
	}

	presignClient := s3.NewPresignClient(c.client)
	url, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(in.BucketName),
		Key:    aws.String(fileName),
	}, s3.WithPresignExpires(constants.URLExpiration*time.Minute))
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
		Url:         url.URL,
		ContentType: in.ContentType,
	}, nil
}

func (c *s3Blob) SignURL(ctx context.Context, in *blob.SignURLRequest) (*blob.SignURLResponse, error) {
	presignClient := s3.NewPresignClient(c.client)
	url, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(in.BucketName),
		Key:    aws.String(in.FileName),
	}, s3.WithPresignExpires(constants.URLExpiration*time.Minute))
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to generate signed URL: %s", err),
		}
	}

	return &blob.SignURLResponse{
		Url: url.URL,
	}, nil
}

func (c *s3Blob) RemoveFile(ctx context.Context, in *blob.DeleteRequest) error {
	if in.BucketName == "" {
		in.BucketName = c.BucketNameImage
	}

	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(in.BucketName),
		Key:    aws.String(in.FileName),
	})
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to delete file %q: %s", in.FileName, err),
		}
	}
	return nil
}
