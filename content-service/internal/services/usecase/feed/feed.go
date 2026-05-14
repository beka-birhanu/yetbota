package feed

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
	feedDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	"golang.org/x/sync/errgroup"
)

func (s *svc) ListFeed(ctx context.Context, ctxSess *ctxRP.Context, req *ListFeedRequest) (*ListFeedResponse, error) {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	count, err := s.feedRepo.Count(ctx, ctxSess.UserSession.UserID)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}
	if count == 0 {
		// TODO: trigger background feed update
		return &ListFeedResponse{}, nil
	}

	fetchOpts := &feedDomain.ListOptions{Limit: req.PageSize}
	if req.Cursor != "" {
		fetchOpts.MaxScore, err = parseCursor(req.Cursor)
		if err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return nil, err
		}
	}

	unseenIDs, nextCursor, err := s.collectUnseenIDs(ctx, ctxSess.UserSession.UserID, fetchOpts, req.PageSize)
	if err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	var posts []*dbmodels.Post
	var photos dbmodels.PostPhotoSlice

	errGrp, egCtx := errgroup.WithContext(ctx)
	errGrp.Go(func() error {
		var e error
		posts, e = s.postRepo.List(egCtx, &post.ListOptions{IDs: unseenIDs, PageSize: len(unseenIDs), Page: 1})
		return e
	})
	errGrp.Go(func() error {
		var e error
		photos, e = s.postPhotoRepo.List(egCtx, &domainPostphoto.Options{PostIDs: unseenIDs, LoadPhoto: true},
			&domainPostphoto.SortOptions{Field: domainPostphoto.SortFieldPosition, Direction: domainPostphoto.SortDirectionAsc})
		return e
	})
	if err = errGrp.Wait(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return nil, err
	}

	return &ListFeedResponse{
		Posts:      orderPosts(posts, unseenIDs),
		Photos:     groupPhotosByPost(photos),
		NextCursor: nextCursor,
	}, nil
}

func (s *svc) MarkViewed(ctx context.Context, ctxSess *ctxRP.Context, req *MarkViewedRequest) error {
	if err := req.Validate(); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	userID := ctxSess.UserSession.UserID

	if err := s.seenRepo.AddBulk(ctx, userID, req.PostIDs); err != nil {
		ctxSess.SetErrorMessage(err.Error())
		return err
	}

	for _, postID := range req.PostIDs {
		key := seenFeedKey(userID, postID)
		if err := s.seenCache.Add(ctx, key, s.seenCacheTTL); err != nil {
			ctxSess.SetErrorMessage(err.Error())
			return err
		}
	}
	return nil
}
