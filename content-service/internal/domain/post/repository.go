package post

import (
	"context"
	"database/sql"
	"errors"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

var ErrConflict = errors.New("optimistic lock conflict")

type Repository interface {
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error
	Read(ctx context.Context, id string) (*dbmodels.Post, error)
	Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error

	GetVote(ctx context.Context, userID, postID string) (*dbmodels.PostVote, error)
	AddVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error
	UpdateVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error
	UpdateCounts(ctx context.Context, tx *sql.Tx, id string, likesDelta, dislikesDelta, expectedLikes, expectedDislikes int) error
}
