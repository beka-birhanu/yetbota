package user

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/internal/domain/storage"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
	"github.com/nyaruka/phonenumbers"
	"golang.org/x/sync/errgroup"
)

func normalizePhone(mobile string) (string, error) {
	parsed, err := phonenumbers.Parse(mobile, constants.DefaultPhoneRegion)
	if err != nil {
		return "", err
	}
	if !phonenumbers.IsValidNumber(parsed) {
		return "", &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			PublicMessage:     "Invalid phone number",
			ServiceStatusCode: status.BadRequestMissingField,
			ServiceMessage:    "invalid phone number",
		}
	}
	return phonenumbers.Format(parsed, phonenumbers.E164), nil
}

func applyUserSelfUpdate(u *domainUser.User, req *UpdateSelfRequest) {
	if req.FirstName != "" {
		u.FirstName = req.FirstName
	}
	if req.LastName != "" {
		u.LastName = req.LastName
	}
	if req.Username != "" {
		u.Username = req.Username
	}
}

func (s *svc) assembleUserWrappers(ctx context.Context, users dbmodels.UserSlice) ([]*UserWrapper, error) {
	wrappers := make([]*UserWrapper, len(users))
	eg, egCtx := errgroup.WithContext(ctx)
	for i, user := range users {
		eg.Go(func() error {
			wrapper, err := s.assembleUserWrapper(egCtx, user)
			if err != nil {
				return err
			}
			wrappers[i] = wrapper
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return wrappers, nil
}

func (s *svc) assembleUserWrapper(ctx context.Context, user *dbmodels.User) (*UserWrapper, error) {
	var photo *dbmodels.Photo
	var err error

	if user.R != nil && user.R.ProfilePhoto != nil {
		photo = user.R.ProfilePhoto
	} else {
		photo, err = s.photoRepo.Read(ctx, user.ProfilePhotoID.String)
		if err != nil {
			return nil, err
		}
	}

	resp, err := s.bucket.SignURL(ctx, &storage.SignURLRequest{
		BucketName: s.bucketName,
		FileName:   photo.URL,
	})
	if err != nil {
		return nil, err
	}

	return &UserWrapper{
		User:       user,
		ProfileURL: resp.Url,
	}, nil
}
