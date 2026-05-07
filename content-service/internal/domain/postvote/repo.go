package postvote

import (
	"context"
	"database/sql"

	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
)

type Repository interface {
	Exists(ctx context.Context, userID, postID string) (bool, error)
	Get(ctx context.Context, userID, postID string) (*dbmodels.PostVote, error)
	Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error
	Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error
	UpdateCounts(ctx context.Context, tx *sql.Tx, id string, likesDelta, dislikesDelta, expectedLikes, expectedDislikes int) error
}
