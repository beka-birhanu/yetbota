package user

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
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

func pickPhotoURL(photo *dbmodels.Photo, res PhotoResolution) string {
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

func (s *svc) assembleUserWrappers(ctx context.Context, users dbmodels.UserSlice, resolution PhotoResolution) ([]*UserWrapper, error) {
	wrappers := make([]*UserWrapper, len(users))
	eg, egCtx := errgroup.WithContext(ctx)
	for i, user := range users {
		eg.Go(func() error {
			wrapper, err := s.assembleUserWrapper(egCtx, user, resolution)
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

func (s *svc) assembleUserWrapper(ctx context.Context, user *dbmodels.User, resolution PhotoResolution) (*UserWrapper, error) {
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

	return &UserWrapper{
		User:       user,
		ProfileURL: pickPhotoURL(photo, resolution),
	}, nil
}
