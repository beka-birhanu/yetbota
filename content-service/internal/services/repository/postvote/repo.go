package postvote

import (
	"context"
	"database/sql"

	"github.com/aarondl/sqlboiler/v4/boil"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
)

type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func NewRepo(c *Config) (postvote.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: c.DB}, nil
}

// Exists implements [postvote.Repository].
func (r *repository) Exists(ctx context.Context, userID string, postID string) (bool, error) {
	exists, err := dbmodels.PostVoteExists(ctx, r.db, userID, postID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Add implements [postvote.Repository].
func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return nil
}

// Get implements [postvote.Repository].
func (r *repository) Get(ctx context.Context, userID string, postID string) (*dbmodels.PostVote, error) {
	v, err := dbmodels.FindPostVote(ctx, r.db, userID, postID)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return v, nil
}

// UpdateCounts implements [postvote.Repository].
func (r *repository) UpdateCounts(ctx context.Context, tx *sql.Tx, id string, likesDelta int, dislikesDelta int, expectedLikes int, expectedDislikes int) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	result, err := exec.ExecContext(
		ctx,
		`
		UPDATE posts
		SET
			likes = likes + $1,
			dislikes = dislikes + $2
		WHERE
			id = $3
			AND likes = $4
			AND dislikes = $5
		`,
		likesDelta,
		dislikesDelta,
		id,
		expectedLikes,
		expectedDislikes,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.Conflict,
			ServiceStatusCode: status.Conflict,
			PublicMessage:     "Please try again later.",
			ServiceMessage:    "optimistic locking failed",
		}
	}

	return nil
}

// Update implements [postvote.Repository].
func (r *repository) Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	rowAff, err := entity.Update(ctx, exec, boil.Infer())
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	if rowAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "post vote not found",
			ServiceMessage:    "post vote not found",
		}
	}
	return nil
}
