package comment

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

var ErrConflict = errors.New("optimistic lock conflict")

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Comment) error
	Read(ctx context.Context, id string) (*dbmodels.Comment, error)
	List(ctx context.Context, opts *Options) (dbmodels.CommentSlice, error)
	Delete(ctx context.Context, tx *sql.Tx, id string) error

	GetVote(ctx context.Context, userID, commentID string) (*dbmodels.CommentVote, error)
	AddVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error
	UpdateVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error
	UpdateCounts(ctx context.Context, tx *sql.Tx, id string, upvoteDelta, downvoteDelta, expectedUpvote, expectedDownvote int) error
}

type Options struct {
	PostID    string
	CommentID string
}
