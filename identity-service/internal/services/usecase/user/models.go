package user

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
)

// List
type ListRequest struct {
	Options    *domainUser.Options    `validate:"required"`
	Pagination *domainUser.Pagination `validate:"required"`
	Sort       *domainUser.SortOption `validate:"required"`
}

func (r *ListRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type PaginationInfo struct {
	Total       int64
	Limit       int
	CurrentPage int
}

// UserWrapper wraps a user with a signed profile URL.
type UserWrapper struct {
	User       *dbmodels.User
	ProfileURL string
}

type ListResponse struct {
	Users      []*UserWrapper
	Pagination *PaginationInfo
}

// Read

type ReadRequest struct {
	ID string `validate:"required"`
}

func (r *ReadRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ReadResponse struct {
	*UserWrapper
}

// ReadPublic

type ReadPublicRequest struct {
	ID string `validate:"required"`
}

func (r *ReadPublicRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ReadPublicResponse struct {
	*UserWrapper
}

// Update

type UpdateRequest struct {
	ID     string `validate:"required"`
	Status string
	Role   string
}

func (r *UpdateRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UpdateResponse struct {
	User *domainUser.User
}

// UpdateSelf

type UpdateSelfRequest struct {
	FirstName string
	LastName  string
	Username  string
}

func (r *UpdateSelfRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UpdateSelfResponse struct {
	User *domainUser.User
}

// Register

type RegisterRequest struct {
	ID        string `validate:"required"`
	FirstName string `validate:"required"`
	LastName  string `validate:"required"`
	Username  string `validate:"required"`
	Mobile    string `validate:"required"`
	Password  string `validate:"required" mask:"true"`
	Random    string `validate:"required" mask:"true"`
}

func (r *RegisterRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	mobile, err := normalizePhone(r.Mobile)
	if err != nil {
		return err
	}
	r.Mobile = mobile
	return nil
}

type RegisterResponse struct {
	User *domainUser.User
}

// Delete

type DeleteRequest struct {
	ID string `validate:"required"`
}

func (r *DeleteRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type DeleteResponse struct{}

// DeleteSelf

type DeleteSelfRequest struct{}

func (r *DeleteSelfRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type DeleteSelfResponse struct{}

// UploadProfile

type UploadProfileRequest struct {
	Image []byte `validate:"required"`
}

func (r *UploadProfileRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type UploadProfileResponse struct {
	URL string
}

// CheckMobile

type CheckMobileRequest struct {
	Mobile string `validate:"required"`
}

func (r *CheckMobileRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	mobile, err := normalizePhone(r.Mobile)
	if err != nil {
		return err
	}
	r.Mobile = mobile
	return nil
}

type CheckMobileResponse struct {
	Exists bool
}

// AddDevice

type AddDeviceRequest struct {
	DeviceID  string
	Token     string
	Oem       string
	Device    string
	OS        string
	Longitude float32
	Latitude  float32
}

type AddDeviceResponse struct{}
