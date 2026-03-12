package auth

import (
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
)

// Login

type LoginRequest struct {
	Username string `validate:"required"`
	Password string `validate:"required" mask:"true"`
	Site     string
}

func (r *LoginRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type LoginResponse struct {
	AccessToken     string
	AccessTokenTTL  int64 // seconds
	RefreshToken    string
	RefreshTokenTTL int64 // seconds
}

// Refresh

type RefreshRequest struct {
	RefreshToken string `validate:"required" mask:"true"`
	Username     string `validate:"required"`
}

func (r *RefreshRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type RefreshResponse struct {
	AccessToken     string
	AccessTokenTTL  int64
	RefreshToken    string
	RefreshTokenTTL int64
}

// Logout

type LogoutRequest struct {
	RefreshToken string `validate:"required" mask:"true"`
	Username     string `validate:"required"`
}

func (r *LogoutRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type LogoutResponse struct{}

// GenerateMobileOTP

type GenerateMobileOTPRequest struct {
	Mobile string `validate:"required"`
	Random string `validate:"required"`
}

func (r *GenerateMobileOTPRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type GenerateMobileOTPResponse struct {
	OtpReqCount int32
	MaxOtpReq   int32
	OtpErrCount int32
	MaxOtpErr   int32
}

// ValidateOTP

type ValidateOTPRequest struct {
	Otp    string `validate:"required" mask:"true"`
	Mobile string `validate:"required"`
	Random string `validate:"required" mask:"true"`
}

func (r *ValidateOTPRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ValidateOTPResponse struct {
	OtpReqCount int32
	MaxOtpReq   int32
	OtpErrCount int32
	MaxOtpErr   int32
}

// NewPassword

type NewPasswordRequest struct {
	Password string `validate:"required,min=8" mask:"true"`
	Random   string `validate:"required" mask:"true"`
	Username string `validate:"required"`
}

func (r *NewPasswordRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type NewPasswordResponse struct{}

// Authorization

type AuthorizationRequest struct {
	Resource string `validate:"required"`
	Action   string `validate:"required"`
}

func (r *AuthorizationRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type AuthorizationResponse struct{}

// ChangePassword

type ChangePasswordRequest struct {
	CurrentPassword string `validate:"required" mask:"true"`
	NewPassword     string `validate:"required,min=8" mask:"true"`
}

func (r *ChangePasswordRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ChangePasswordResponse struct{}

// ChangeMobile

type ChangeMobileRequest struct {
	NewMobile string `validate:"required"`
	Random    string `validate:"required" mask:"true"`
}

func (r *ChangeMobileRequest) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type ChangeMobileResponse struct{}
