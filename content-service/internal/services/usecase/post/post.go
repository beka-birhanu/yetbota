package post

import (
	"context"
	"errors"
	"sync"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/geotypes"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/twpayne/go-geom"
	"golang.org/x/sync/errgroup"
)

func (s *svc) Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	post := postFromAddReq(req)
	post.UserID = ctxSess.UserSession.UserID

	uploaded, err := s.uploadPhotos(ctx, post.ID, req.Photos)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	defer func() {
		if err != nil {
			err = s.deleteUploads(ctx, uploaded.photos)
			if err != nil {
				ctxSess.SetErrorMessage(ctxSess.ErrorMessage + "\n" + err.Error())
			}
		}
	}()

	tx, err := repository.BeginNewTx(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = s.postRepo.Add(ctx, tx, post)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	err = s.photoRepo.AddBulk(ctx, tx, uploaded.photos)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	err = s.postPhotoRepo.AddBulk(ctx, tx, uploaded.postPhotos)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	orderedPhotos, err := s.assembleOrderedPhoto(ctx, uploaded.postPhotos, PhotoResolutionOriginal)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &AddResponse{Post: post, Photos: orderedPhotos}, nil
}

func (s *svc) Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	post, err := s.postRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	photos, err := s.postPhotoRepo.List(ctx, &domainPostphoto.Options{
		PostID:    post.ID,
		LoadPhoto: true,
	}, &domainPostphoto.SortOptions{
		Field:     domainPostphoto.SortFieldPosition,
		Direction: domainPostphoto.SortDirectionAsc,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	orderedPhotos, err := s.assembleOrderedPhoto(ctx, photos, req.PhotoResolution)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ReadResponse{Post: post, Photos: orderedPhotos}, nil
}

func (s *svc) Update(ctx context.Context, ctxSess *ctxRP.Context, req *UpdateRequest) (*UpdateResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	post, err := s.postRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var uploaded *uploadPhotosResponse
	var positionMap map[int]*dbmodels.PostPhoto
	var oldPhotoURLs []string

	if len(req.UpsertPhotos) > 0 {
		uploaded, err = s.uploadPhotos(ctx, post.ID, req.UpsertPhotos)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		defer func() {
			if err != nil {
				_ = s.deleteUploads(ctx, uploaded.photos)
			}
		}()

		existing, listErr := s.postPhotoRepo.List(ctx, &domainPostphoto.Options{
			PostID:    post.ID,
			LoadPhoto: true,
		}, nil)
		if listErr != nil {
			err = listErr
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		positionMap = make(map[int]*dbmodels.PostPhoto, len(existing))
		for _, pp := range existing {
			positionMap[pp.Position] = pp
		}
	}

	tx, err := repository.BeginNewTx(ctx)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	post.Title = req.Title
	post.Description = req.Description
	post.Tags = req.Tags
	post.Location = geotypes.NullPoint{Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{req.Latitude, req.Longitude})}
	post.Address = null.NewString(req.Address, req.Address != "")

	err = s.postRepo.Update(ctx, tx, post)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if len(req.UpsertPhotos) > 0 {
		toInsertPhotos := make(dbmodels.PhotoSlice, 0)
		toInsertPostPhotos := make(dbmodels.PostPhotoSlice, 0)

		for i, newPP := range uploaded.postPhotos {
			newPhoto := uploaded.photos[i]

			if existingPP, exists := positionMap[newPP.Position]; exists {
				if existingPP.R != nil && existingPP.R.Photo != nil {
					oldPhotoURLs = append(oldPhotoURLs, existingPP.R.Photo.URL)
				}
				oldPhotoID := existingPP.PhotoID

				err = s.photoRepo.Add(ctx, tx, newPhoto)
				if err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}

				existingPP.PhotoID = newPhoto.ID
				err = s.postPhotoRepo.Update(ctx, tx, existingPP)
				if err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}

				err = s.photoRepo.Delete(ctx, tx, oldPhotoID)
				if err != nil {
					ctxSess.SetErrorMessage(err.Error())
					return nil, err
				}
			} else {
				toInsertPhotos = append(toInsertPhotos, newPhoto)
				toInsertPostPhotos = append(toInsertPostPhotos, newPP)
			}
		}

		if len(toInsertPhotos) > 0 {
			err = s.photoRepo.AddBulk(ctx, tx, toInsertPhotos)
			if err != nil {
				ctxSess.SetErrorMessage(err.Error())
				return nil, err
			}
			err = s.postPhotoRepo.AddBulk(ctx, tx, toInsertPostPhotos)
			if err != nil {
				ctxSess.SetErrorMessage(err.Error())
				return nil, err
			}
		}
	}

	if err = repository.CommitTx(tx); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if len(oldPhotoURLs) > 0 {
		oldPhotos := make(dbmodels.PhotoSlice, 0, len(oldPhotoURLs))
		for _, url := range oldPhotoURLs {
			oldPhotos = append(oldPhotos, &dbmodels.Photo{URL: url})
		}
		_ = s.deleteUploads(ctx, oldPhotos)
	}

	return &UpdateResponse{Post: post}, nil
}

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	posts, total, err := s.postRepo.List(ctx, &req.ListOptions)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	photosByPost := make(map[string][]*OrderedPhoto, len(posts))
	eg, egCtx := errgroup.WithContext(ctx)
	var mu sync.Mutex

	for _, p := range posts {
		postID := p.ID
		eg.Go(func() error {
			ppList, err := s.postPhotoRepo.List(egCtx, &domainPostphoto.Options{
				PostID:    postID,
				LoadPhoto: true,
			}, &domainPostphoto.SortOptions{
				Field:     domainPostphoto.SortFieldPosition,
				Direction: domainPostphoto.SortDirectionAsc,
			})
			if err != nil {
				return err
			}
			ordered, err := s.assembleOrderedPhoto(egCtx, ppList, req.PhotoResolution)
			if err != nil {
				return err
			}
			mu.Lock()
			photosByPost[postID] = ordered
			mu.Unlock()
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ListResponse{
		Posts:    posts,
		Photos:   photosByPost,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *svc) Vote(ctx context.Context, ctxSess *ctxRP.Context, req *PostVoteRequest) (*PostVoteResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	existingVote, err := s.postRepo.GetVote(ctx, ctxSess.UserSession.UserID, req.PostID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if existingVote != nil && existingVote.VoteType == req.VoteType {
		post, err := s.postRepo.Read(ctx, req.PostID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		return &PostVoteResponse{Likes: post.Likes, Dislikes: post.Dislikes}, nil
	}

	voteEntity := &dbmodels.PostVote{
		UserID:   ctxSess.UserSession.UserID,
		PostID:   req.PostID,
		VoteType: req.VoteType,
	}

	const maxRetries = 3
	for range maxRetries {
		post, err := s.postRepo.Read(ctx, req.PostID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		var likesDelta, dislikesDelta int
		switch req.VoteType {
		case dbmodels.PostVoteTypeLike:
			likesDelta = 1
			if existingVote != nil {
				dislikesDelta = -1
			}
		case dbmodels.PostVoteTypeDislike:
			dislikesDelta = 1
			if existingVote != nil {
				likesDelta = -1
			}
		}

		tx, err := repository.BeginNewTx(ctx)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		if existingVote == nil {
			err = s.postRepo.AddVote(ctx, tx, voteEntity)
		} else {
			err = s.postRepo.UpdateVote(ctx, tx, voteEntity)
		}
		if err != nil {
			_ = tx.Rollback()
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		err = s.postRepo.UpdateCounts(ctx, tx, req.PostID, likesDelta, dislikesDelta, post.Likes, post.Dislikes)
		if errors.Is(err, domainPost.ErrConflict) {
			_ = tx.Rollback()
			continue
		}
		if err != nil {
			_ = tx.Rollback()
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		if err = repository.CommitTx(tx); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		return &PostVoteResponse{Likes: post.Likes + likesDelta, Dislikes: post.Dislikes + dislikesDelta}, nil
	}

	err = &toddlerr.Error{
		PublicStatusCode:  status.Conflict,
		ServiceStatusCode: status.Conflict,
		PublicMessage:     "too many concurrent updates, please try again",
		ServiceMessage:    "max vote retries exceeded for post " + req.PostID,
	}
	ctxSess.SetErrorMessage(err.Error())
	return nil, err
}
