package comment

import (
	"context"
	"errors"
	"fmt"

	"github.com/aarondl/null/v8"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/utils"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	repository "github.com/beka-birhanu/yetbota/content-service/internal/services/repository"
	"github.com/google/uuid"
)

const maxOptLockRetries = 3

func (s *svc) Add(ctx context.Context, ctxSess *ctxRP.Context, req *AddRequest) (*AddResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	comment := &dbmodels.Comment{
		ID:        uuid.New().String(),
		Comment:   req.Comment,
		UserID:    ctxSess.UserSession.UserID,
		PostID:    req.PostID,
		CommentID: null.NewString(req.CommentID, req.CommentID != ""),
		IsAnswer:  req.IsAnswer,
	}

	if err := s.commentRepo.Add(ctx, nil, comment); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &AddResponse{Comment: comment}, nil
}

func (s *svc) Read(ctx context.Context, ctxSess *ctxRP.Context, req *ReadRequest) (*ReadResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	comment, err := s.commentRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ReadResponse{Comment: comment}, nil
}

func (s *svc) List(ctx context.Context, ctxSess *ctxRP.Context, req *ListRequest) (*ListResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	comments, err := s.commentRepo.List(ctx, &domainComment.Options{
		PostID:    req.PostID,
		CommentID: req.CommentID,
	})
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ListResponse{Comments: comments}, nil
}

func (s *svc) Delete(ctx context.Context, ctxSess *ctxRP.Context, req *DeleteRequest) error {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	comment, err := s.commentRepo.Read(ctx, req.ID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if comment.UserID != ctxSess.UserSession.UserID {
		err = &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			PublicMessage:     "Either the comment does not exist or you don't have access to delete it.",
			ServiceStatusCode: status.Forbidden,
			ServiceMessage:    fmt.Sprintf("user %s tried to delete a comment under the ownership of %s", ctxSess.UserSession.UserID, comment.UserID),
		}
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	if err := s.commentRepo.Delete(ctx, nil, req.ID); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	return nil
}

func (s *svc) Vote(ctx context.Context, ctxSess *ctxRP.Context, req *VoteRequest) (*VoteResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if err := utils.AllowAccess(ctxSess); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	existingVote, err := s.commentRepo.GetVote(ctx, ctxSess.UserSession.UserID, req.CommentID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	if existingVote != nil && existingVote.VoteType == req.VoteType {
		comment, err := s.commentRepo.Read(ctx, req.CommentID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
		return &VoteResponse{Upvote: comment.Upvote, Downvote: comment.Downvote}, nil
	}

	voteEntity := &dbmodels.CommentVote{
		UserID:    ctxSess.UserSession.UserID,
		CommentID: req.CommentID,
		VoteType:  req.VoteType,
	}

	for range maxOptLockRetries {
		comment, err := s.commentRepo.Read(ctx, req.CommentID)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		var upvoteDelta, downvoteDelta int
		switch req.VoteType {
		case dbmodels.CommentVoteTypeUpvote:
			upvoteDelta = 1
			if existingVote != nil {
				downvoteDelta = -1
			}
		case dbmodels.CommentVoteTypeDownvote:
			downvoteDelta = 1
			if existingVote != nil {
				upvoteDelta = -1
			}
		}

		tx, err := repository.BeginNewTx(ctx)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		if existingVote == nil {
			err = s.commentRepo.AddVote(ctx, tx, voteEntity)
		} else {
			err = s.commentRepo.UpdateVote(ctx, tx, voteEntity)
		}
		if err != nil {
			_ = tx.Rollback()
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}

		err = s.commentRepo.UpdateCounts(ctx, tx, req.CommentID, upvoteDelta, downvoteDelta, comment.Upvote, comment.Downvote)
		if errors.Is(err, domainComment.ErrConflict) {
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

		return &VoteResponse{Upvote: comment.Upvote + upvoteDelta, Downvote: comment.Downvote + downvoteDelta}, nil
	}

	err = &toddlerr.Error{
		PublicStatusCode:  status.Conflict,
		ServiceStatusCode: status.Conflict,
		PublicMessage:     "too many concurrent updates, please try again",
		ServiceMessage:    "max vote retries exceeded for comment " + req.CommentID,
	}
	ctxSess.SetErrorMessage(err.Error())
	return nil, err
}
