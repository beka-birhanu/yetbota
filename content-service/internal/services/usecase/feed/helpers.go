package feed

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	feedDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	postSvc "github.com/beka-birhanu/yetbota/content-service/internal/services/usecase/post"
)

// collectUnseenIDs fetches feed items for userID, filters out seen ones, and returns
// up to pageSize unseen post IDs plus a cursor for the next page.
func (s *svc) collectUnseenIDs(ctx context.Context, userID string, opts *feedDomain.ListOptions, pageSize int) ([]string, string, error) {
	ids := make([]string, pageSize)
	size := 0
	var lastFeedItem *feedDomain.FeedItem

	for {
		items, err := s.feedRepo.List(ctx, userID, opts)
		if err != nil {
			return nil, "", err
		}

		keys := make([]string, len(items))
		for i, item := range items {
			keys[i] = seenFeedKey(userID, item.PostID)
		}

		seenMap, err := s.seenCache.Exists(ctx, keys)
		if err != nil {
			return nil, "", err
		}

		size = 0
		for _, item := range items {
			if seenMap[seenFeedKey(userID, item.PostID)] {
				continue
			}
			ids[size] = item.PostID
			lastFeedItem = item
			size++
			if size == pageSize {
				break
			}
		}

		if size == pageSize || len(items) < opts.Limit {
			break
		}
		opts.Limit *= 2
	}

	var nextCursor string
	if size >= pageSize {
		nextCursor = buildNextCursor(lastFeedItem)
	}

	return ids[:size], nextCursor, nil
}

func orderPosts(unordered []*dbmodels.Post, ids []string) []*dbmodels.Post {
	byID := make(map[string]*dbmodels.Post, len(unordered))
	for _, p := range unordered {
		byID[p.ID] = p
	}
	ordered := make([]*dbmodels.Post, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			ordered = append(ordered, p)
		}
	}
	return ordered
}

func groupPhotosByPost(photos dbmodels.PostPhotoSlice) map[string][]*postSvc.OrderedPhoto {
	m := make(map[string][]*postSvc.OrderedPhoto)
	for _, pp := range photos {
		var photoURL string
		if pp.R != nil && pp.R.Photo != nil {
			photoURL = pp.R.Photo.URL
		}
		m[pp.PostID] = append(m[pp.PostID], &postSvc.OrderedPhoto{
			ID:       pp.PhotoID,
			PostID:   pp.PostID,
			URL:      photoURL,
			Position: pp.Position,
		})
	}
	return m
}

// buildNextCursor encodes score and postID so the next page can start after this item.
func buildNextCursor(item *feedDomain.FeedItem) string {
	if item == nil {
		return ""
	}
	return fmt.Sprintf("cursor:%g", item.Score)
}

// parseCursor decodes a cursor string into (maxScore, afterPostID).
func parseCursor(cursor string) (float64, error) {
	if cursor == "" {
		return 0, nil
	}

	scoreStr := strings.TrimPrefix(cursor, "cursor:")
	score, err := strconv.ParseFloat(scoreStr, 64)
	if err != nil {
		return 0, &toddlerr.Error{
			PublicStatusCode:  status.BadRequest,
			ServiceStatusCode: status.BadRequest,
			PublicMessage:     "Invalid Request",
			ServiceMessage:    fmt.Sprintf("Invalid cursor: %v", err),
		}
	}

	return score, nil
}

// seenFeedKey builds a key for the seen feed DB table.
func seenFeedKey(userID, postID string) string {
	return fmt.Sprintf("%s:%s", userID, postID)
}
