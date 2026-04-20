package auth

import (
	"context"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainAuth "github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	contextYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	"github.com/beka-birhanu/yetbota/identity-service/internal/domain/messaging"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
)

type Service interface {
	Login(ctx context.Context, ctxSess *contextYB.Context, req *LoginRequest) (*LoginResponse, error)
	Refresh(ctx context.Context, ctxSess *contextYB.Context, req *RefreshRequest) (*RefreshResponse, error)
	Logout(ctx context.Context, ctxSess *contextYB.Context, req *LogoutRequest) (*LogoutResponse, error)
	GenerateMobileOTP(ctx context.Context, ctxSess *contextYB.Context, req *GenerateMobileOTPRequest) (*GenerateMobileOTPResponse, error)
	ValidateOTP(ctx context.Context, ctxSess *contextYB.Context, req *ValidateOTPRequest) (*ValidateOTPResponse, error)
	NewPassword(ctx context.Context, ctxSess *contextYB.Context, req *NewPasswordRequest) (*NewPasswordResponse, error)
	Authorization(ctx context.Context, ctxSess *contextYB.Context, req *AuthorizationRequest) (*AuthorizationResponse, error)
	ChangePassword(ctx context.Context, ctxSess *contextYB.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error)
	ChangeMobile(ctx context.Context, ctxSess *contextYB.Context, req *ChangeMobileRequest) (*ChangeMobileResponse, error)
}

type Config struct {
	UserRepo       domainUser.Repository     `validate:"required"`
	OtpStore       domainAuth.OtpStore       `validate:"required"`
	SessionManager domainAuth.SessionManager `validate:"required"`
	Hasher         domainAuth.Hasher         `validate:"required"`
	SMSClient      messaging.SMSClient       `validate:"required"`
	OtpTTL         time.Duration             `validate:"required"`
	LockRequestTTL time.Duration             `validate:"required"`
	LockInvalidTTL time.Duration             `validate:"required"`
	AccessTTL      time.Duration             `validate:"required"`
	RefreshTTL     time.Duration             `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type service struct {
	userRepo       domainUser.Repository
	otpStore       domainAuth.OtpStore
	sessionManager domainAuth.SessionManager
	hasher         domainAuth.Hasher
	smsClient      messaging.SMSClient
	otpTTL         time.Duration
	lockRequestTTL time.Duration
	lockInvalidTTL time.Duration
	accessTTL      time.Duration
	refreshTTL     time.Duration
}

func NewService(cfg *Config) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &service{
		userRepo:       cfg.UserRepo,
		otpStore:       cfg.OtpStore,
		sessionManager: cfg.SessionManager,
		hasher:         cfg.Hasher,
		smsClient:      cfg.SMSClient,
		otpTTL:         cfg.OtpTTL,
		lockRequestTTL: cfg.LockRequestTTL,
		lockInvalidTTL: cfg.LockInvalidTTL,
		accessTTL:      cfg.AccessTTL,
		refreshTTL:     cfg.RefreshTTL,
	}, nil
}
