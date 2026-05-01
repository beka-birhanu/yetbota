package post

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

var ErrConflict = errors.New("optimistic lock conflict")

type ListSortField string

const (
	ListSortFieldCreatedAt ListSortField = "created_at"
	ListSortFieldLikes     ListSortField = "likes"
	ListSortFieldDislikes  ListSortField = "dislikes"
	ListSortFieldComments  ListSortField = "comment_count"
)

type ListSortDir string

const (
	ListSortDirAsc  ListSortDir = "ASC"
	ListSortDirDesc ListSortDir = "DESC"
)

type ListOptions struct {
	UserID     string
	Tags       []string
	IsQuestion *bool
	Search     string
	NearLat    *float64
	NearLon    *float64
	RadiusKm   *float64
	SortField  ListSortField
	SortDir    ListSortDir
	Page       int
	PageSize   int
}

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error
	Read(ctx context.Context, id string) (*dbmodels.Post, error)
	Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error
	List(ctx context.Context, opts *ListOptions) ([]*dbmodels.Post, int64, error)

	GetVote(ctx context.Context, userID, postID string) (*dbmodels.PostVote, error)
	AddVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error
	UpdateVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error
	UpdateCounts(ctx context.Context, tx *sql.Tx, id string, likesDelta, dislikesDelta, expectedLikes, expectedDislikes int) error
}
