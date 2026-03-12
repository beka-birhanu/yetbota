package user

import (
	"context"
	"fmt"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	domainStorage "github.com/beka-birhanu/yetbota/identity-service/internal/domain/storage"
	domainUser "github.com/beka-birhanu/yetbota/identity-service/internal/domain/user"
	"github.com/beka-birhanu/yetbota/identity-service/internal/services/repository"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAdminAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var users []*dbmodels.User
	var count int64

	req.Options.LoadPhoto = true
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		users, err = s.userRepo.List(egCtx, req.Options, req.Pagination, req.Sort)
		return err
	})
	eg.Go(func() error {
		var err error
		count, err = s.userRepo.Count(egCtx, req.Options)
		return err
	})

	if err := eg.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	wrappers, err := s.assembleUserWrappers(ctx, users)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ListResponse{
		Users: wrappers,
		Pagination: &PaginationInfo{
			Total:       count,
			Limit:       req.Pagination.Limit,
			CurrentPage: req.Pagination.Page,
		},
	}, nil
}

func (s *svc) Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	// if user is not self, check if admin
	if err := utils.AllowAdminAccess(ctxSess); ctxSess.UserSession.SessionID != req.ID && err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.Read(ctx, req.ID, &domainUser.Options{LoadPhoto: true})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	wrapper, err := s.assembleUserWrapper(ctx, user)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ReadResponse{UserWrapper: wrapper}, nil
}

func (s *svc) ReadPublic(ctx context.Context, ctxSess *ctxRP.Context, req *ReadPublicRequest) (*ReadPublicResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.Read(ctx, req.ID, &domainUser.Options{LoadPhoto: true})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	wrapper, err := s.assembleUserWrapper(ctx, user)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ReadPublicResponse{UserWrapper: wrapper}, nil
}

func (s *svc) Update(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateRequest) (*UpdateResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAdminAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user, err := s.userRepo.Read(ctx, req.ID, nil)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if req.Status != "" {
		user.Status = req.Status
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	cols := boil.Whitelist(dbmodels.UserColumns.Status, dbmodels.UserColumns.Role)
	if err := s.userRepo.Update(ctx, nil, user, cols); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &UpdateResponse{User: user}, nil
}

func (s *svc) UpdateSelf(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateSelfRequest) (*UpdateSelfResponse, error) {
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

	applyUserSelfUpdate(user, req)

	cols := boil.Whitelist(dbmodels.UserColumns.FirstName, dbmodels.UserColumns.LastName, dbmodels.UserColumns.Username)
	if err := s.userRepo.Update(ctx, nil, user, cols); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &UpdateSelfResponse{User: user}, nil
}

func (s *svc) Register(ctx context.Context, ctxSess *ctxRP.Context, req *RegisterRequest) (*RegisterResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	key := "otp:mobile:" + req.Mobile

	otp, err := s.otpStore.Read(ctx, key)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if otp.Otp == "" || !otp.Verified || s.hasher.Verify(otp.Random, req.Random) != nil {
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "OTP not verified",
			ServiceMessage:    fmt.Sprintf("otp not verified for mobile: %s", req.Mobile),
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	hashed, err := s.hasher.Hash(req.Password)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	newUser := &dbmodels.User{
		ID:        req.ID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Username:  req.Username,
		Mobile:    req.Mobile,
		Password:  hashed,
	}

	created, err := s.userRepo.Add(ctx, nil, newUser)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &RegisterResponse{User: created}, nil
}

func (s *svc) Delete(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAdminAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if _, err := s.userRepo.Read(ctx, req.ID, nil); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.userRepo.Delete(ctx, nil, req.ID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &DeleteResponse{}, nil
}

func (s *svc) DeleteSelf(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteSelfRequest) (*DeleteSelfResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.userRepo.Delete(ctx, nil, ctxSess.UserSession.UserID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &DeleteSelfResponse{}, nil
}

func (s *svc) UploadProfile(ctx context.Context, ctxSess *ctxRP.Context, req *UploadProfileRequest) (*UploadProfileResponse, error) {
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

	if _, err := s.userRepo.Read(ctx, ctxSess.UserSession.UserID, nil); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if int64(len(req.Image)) > constants.MaxUploadSize {
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     fmt.Sprintf("Image exceeds maximum size of %dMB", constants.MaxUploadSize/constants.MB),
			ServiceMessage:    "image too large",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	processed, mime, err := utils.ProcessImage(req.Image)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	uploadResp, err := s.bucket.UploadFile(ctx, &domainStorage.UploadRequest{
		BucketName:  s.bucketName,
		FileInByte:  processed,
		ContentType: mime,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	signResp, err := s.bucket.SignURL(ctx, &domainStorage.SignURLRequest{
		BucketName: s.bucketName,
		FileName:   uploadResp.FileName,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	tx, err := repository.BeginNewTx(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			ctxSess.Logger.Error(ctx, err.Error())
		}
	}()

	// Store object key in URL so we can generate signed URLs on read/list.
	photo, err := s.photoRepo.Add(ctx, tx, &dbmodels.Photo{
		ID:             uuid.NewString(),
		BucketProvider: dbmodels.PhotoBucketS3,
		MimeType:       uploadResp.ContentType,
		URL:            uploadResp.FileName,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	user.ProfilePhotoID = null.StringFrom(photo.ID)
	err = s.userRepo.Update(ctx, tx, user, boil.Whitelist(dbmodels.UserColumns.ProfilePhotoID))
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		err = &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Something went wrong",
			ServiceMessage:    fmt.Sprintf("failed to commit transaction: %s", err),
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &UploadProfileResponse{
		URL: signResp.Url,
	}, nil
}

func (s *svc) Follow(ctx context.Context, ctxSess *ctxRP.Context, req *FollowRequest) (*FollowResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	followerID := ctxSess.UserSession.UserID
	if followerID == req.FolloweeID {
		err := &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Cannot follow yourself",
			ServiceMessage:    "user attempted to follow themselves",
		}
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := s.followRepo.Follow(ctx, followerID, req.FolloweeID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var follower, followee *dbmodels.User
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		follower, err = s.userRepo.Read(egCtx, followerID, nil)
		return err
	})
	eg.Go(func() error {
		var err error
		followee, err = s.userRepo.Read(egCtx, req.FolloweeID, nil)
		return err
	})
	if err := eg.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	follower.Following++
	followee.Followers++

	eg2, eg2Ctx := errgroup.WithContext(ctx)
	eg2.Go(func() error {
		return s.userRepo.Update(eg2Ctx, nil, follower, boil.Whitelist(dbmodels.UserColumns.Following))
	})
	eg2.Go(func() error {
		return s.userRepo.Update(eg2Ctx, nil, followee, boil.Whitelist(dbmodels.UserColumns.Followers))
	})
	if err := eg2.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &FollowResponse{}, nil
}

func (s *svc) Unfollow(ctx context.Context, ctxSess *ctxRP.Context, req *UnfollowRequest) (*UnfollowResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	followerID := ctxSess.UserSession.UserID

	if err := s.followRepo.Unfollow(ctx, followerID, req.FolloweeID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var follower, followee *dbmodels.User
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		follower, err = s.userRepo.Read(egCtx, followerID, nil)
		return err
	})
	eg.Go(func() error {
		var err error
		followee, err = s.userRepo.Read(egCtx, req.FolloweeID, nil)
		return err
	})
	if err := eg.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if follower.Following > 0 {
		follower.Following--
	}
	if followee.Followers > 0 {
		followee.Followers--
	}

	eg2, eg2Ctx := errgroup.WithContext(ctx)
	eg2.Go(func() error {
		return s.userRepo.Update(eg2Ctx, nil, follower, boil.Whitelist(dbmodels.UserColumns.Following))
	})
	eg2.Go(func() error {
		return s.userRepo.Update(eg2Ctx, nil, followee, boil.Whitelist(dbmodels.UserColumns.Followers))
	})
	if err := eg2.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &UnfollowResponse{}, nil
}

func (s *svc) CheckMobile(ctx context.Context, ctxSess *ctxRP.Context, req *CheckMobileRequest) (*CheckMobileResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	exists, err := s.userRepo.MobileExists(ctx, req.Mobile)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &CheckMobileResponse{Exists: exists}, nil
}
