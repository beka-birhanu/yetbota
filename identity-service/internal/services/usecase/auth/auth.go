package auth

import (
	"context"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/utils"
	domainAuth "github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	contextYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
)

func (s *service) Login(ctx context.Context, ctxSess *contextYB.Context, req *LoginRequest) (*LoginResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := req.normalize(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.ReadByUsername(ctx, req.Username, nil)
	if err != nil {
		err := invalidCredentialsError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if user.Status != dbmodels.UserStatusACTIVE {
		err := &toddlerr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Invalid username or password",
			ServiceMessage:    "user status is not active",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.hasher.Verify(user.Password, req.Password); err != nil {
		err := invalidCredentialsError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	td, err := s.sessionManager.NewSessionDetails(ctx, &domainAuth.SessionInfo{
		Username:   user.Username,
		UserID:     user.ID,
		Role:       user.Role,
		RefreshTTL: s.refreshTTL,
		AccessTTL:  s.accessTTL,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.sessionManager.SaveSession(ctx, td); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &LoginResponse{
		AccessToken:     td.AccessToken,
		AccessTokenTTL:  int64(td.AccessTtl.Seconds()),
		RefreshToken:    td.RefreshToken,
		RefreshTokenTTL: int64(td.RefreshTtl.Seconds()),
	}, nil
}

func (s *service) Refresh(ctx context.Context, ctxSess *contextYB.Context, req *RefreshRequest) (*RefreshResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	userSess, err := s.sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
		TokenType: domainAuth.RefreshToken,
		Token:     req.RefreshToken,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if userSess.Username != req.Username {
		err := &toddlerr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Invalid refresh token",
			ServiceMessage:    "username mismatch in refresh token",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	deleted, err := s.sessionManager.DeleteSession(ctx, userSess)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if deleted == 0 {
		err := &toddlerr.Error{
			PublicStatusCode:  status.Unauthorized,
			ServiceStatusCode: status.Unauthorized,
			PublicMessage:     "Session expired",
			ServiceMessage:    "refresh session not found in redis",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.ReadByUsername(ctx, req.Username, nil)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	td, err := s.sessionManager.NewSessionDetails(ctx, &domainAuth.SessionInfo{
		Username:   user.Username,
		UserID:     user.ID,
		Role:       user.Role,
		RefreshTTL: s.refreshTTL,
		AccessTTL:  s.accessTTL,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.sessionManager.SaveSession(ctx, td); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &RefreshResponse{
		AccessToken:     td.AccessToken,
		AccessTokenTTL:  int64(td.AccessTtl.Seconds()),
		RefreshToken:    td.RefreshToken,
		RefreshTokenTTL: int64(td.RefreshTtl.Seconds()),
	}, nil
}

func (s *service) Logout(ctx context.Context, ctxSess *contextYB.Context, req *LogoutRequest) (*LogoutResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	userSess, err := s.sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
		TokenType: domainAuth.RefreshToken,
		Token:     req.RefreshToken,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if _, err := s.sessionManager.DeleteSession(ctx, userSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &LogoutResponse{}, nil
}

func (s *service) GenerateMobileOTP(ctx context.Context, ctxSess *contextYB.Context, req *GenerateMobileOTPRequest) (*GenerateMobileOTPResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	mobile, err := normalizePhone(req.Mobile)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	key := mobileOtpKey(mobile)
	existing, err := s.otpStore.Read(ctx, key)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	now := time.Now()

	if existing.LockedUntil.Valid && now.Before(existing.LockedUntil.Time) {
		err := lockedError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if existing.GeneratedCount >= constants.MAXOTP {
		lockUntil := now.Add(s.lockRequestTTL)
		existing.LockedUntil = null.TimeFrom(lockUntil)
		existing.GeneratedCount = 0
		err = s.otpStore.Save(ctx, existing, key, s.lockRequestTTL)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		err := lockedError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	otpCode, err := GenerateOTP(6)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	hashedCode, err := s.hasher.Hash(otpCode)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	hashedRandom, err := s.hasher.Hash(req.Random)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	otp := domainAuth.Otp{
		Otp:            hashedCode,
		Random:         hashedRandom,
		GeneratedCount: existing.GeneratedCount + 1,
		ErrorCount:     0,
		IssuedAt:       now,
		Verified:       false,
	}

	if err := s.otpStore.Save(ctx, otp, key, s.otpTTL); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	err = s.smsClient.SendOTP(ctx, mobile, otpCode)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &GenerateMobileOTPResponse{
		OtpReqCount: otp.GeneratedCount,
		MaxOtpReq:   constants.MAXOTP,
		OtpErrCount: otp.ErrorCount,
		MaxOtpErr:   constants.MaxNotMatchOtp,
	}, nil
}

func (s *service) ValidateOTP(ctx context.Context, ctxSess *contextYB.Context, req *ValidateOTPRequest) (*ValidateOTPResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	mobile, err := normalizePhone(req.Mobile)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	key := mobileOtpKey(mobile)
	otp, err := s.otpStore.Read(ctx, key)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if otp.Otp == "" {
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "OTP not found or expired",
			ServiceMessage:    "otp not found for identifier: " + req.Mobile,
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	now := time.Now()
	if otp.LockedUntil.Valid && now.Before(otp.LockedUntil.Time) {
		err := lockedError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if s.hasher.Verify(otp.Random, req.Random) != nil {
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid OTP session",
			ServiceMessage:    "random mismatch",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if s.hasher.Verify(otp.Otp, req.Otp) != nil {
		otp.ErrorCount++
		if otp.ErrorCount >= constants.MaxNotMatchOtp {
			otp.LockedUntil = null.TimeFrom(now.Add(s.lockInvalidTTL))
			otp.ErrorCount = 0
		}
		err = s.otpStore.Save(ctx, otp, key, s.otpTTL)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid OTP",
			ServiceMessage:    "otp mismatch",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	otp.Verified = true
	err = s.otpStore.Save(ctx, otp, key, s.otpTTL)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ValidateOTPResponse{
		OtpReqCount: otp.GeneratedCount,
		MaxOtpReq:   constants.MAXOTP,
		OtpErrCount: otp.ErrorCount,
		MaxOtpErr:   constants.MaxNotMatchOtp,
	}, nil
}

func (s *service) NewPassword(ctx context.Context, ctxSess *contextYB.Context, req *NewPasswordRequest) (*NewPasswordResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := req.normalize(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.ReadByMobile(ctx, req.Mobile, nil)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	key := mobileOtpKey(user.Mobile)

	otp, err := s.otpStore.Read(ctx, key)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if otp.Otp == "" || !otp.Verified || s.hasher.Verify(otp.Random, req.Random) != nil {
		err := otpNotVerifiedError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	hashed, err := s.hasher.Hash(req.Password)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user.Password = hashed
	if err := s.userRepo.Update(ctx, nil, user, boil.Whitelist(dbmodels.UserColumns.Password)); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &NewPasswordResponse{}, nil
}

func (s *service) Authorization(ctx context.Context, ctxSess *contextYB.Context, req *AuthorizationRequest) (*AuthorizationResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	return &AuthorizationResponse{}, nil
}

func (s *service) ChangePassword(ctx context.Context, ctxSess *contextYB.Context, req *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.Read(ctx, ctxSess.UserSession.UserID, nil)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.hasher.Verify(user.Password, req.CurrentPassword); err != nil {
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Current password is incorrect",
			ServiceMessage:    "current password hash mismatch",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	hashed, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user.Password = hashed
	if err := s.userRepo.Update(ctx, nil, user, boil.Whitelist(dbmodels.UserColumns.Password)); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ChangePasswordResponse{}, nil
}

func (s *service) ChangeMobile(ctx context.Context, ctxSess *contextYB.Context, req *ChangeMobileRequest) (*ChangeMobileResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.Read(ctx, ctxSess.UserSession.UserID, nil)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	newMobile, err := normalizePhone(req.NewMobile)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	key := mobileOtpKey(newMobile)

	otp, err := s.otpStore.Read(ctx, key)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if otp.Otp == "" || !otp.Verified || s.hasher.Verify(otp.Random, req.Random) != nil {
		err := otpNotVerifiedError()
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user.Mobile = newMobile
	if err := s.userRepo.Update(ctx, nil, user, boil.Whitelist(dbmodels.UserColumns.Mobile)); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ChangeMobileResponse{}, nil
}
