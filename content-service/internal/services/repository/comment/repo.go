package comment

import (
	"context"
	"database/sql"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainComment "github.com/beka-birhanu/yetbota/content-service/internal/domain/comment"
)


type repository struct {
	db *sql.DB
}

type Config struct {
	DB *sql.DB `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRepo(cfg *Config) (domainComment.Repository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &repository{db: cfg.DB}, nil
}

func (r *repository) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Comment) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return nil
}

func (r *repository) Read(ctx context.Context, id string) (*dbmodels.Comment, error) {
	c, err := dbmodels.FindComment(ctx, r.db, id)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return c, nil
}

func (r *repository) List(ctx context.Context, opts *domainComment.Options) (dbmodels.CommentSlice, error) {
	mods := []qm.QueryMod{}

	if opts != nil {
		if opts.PostID != "" {
			mods = append(mods, dbmodels.CommentWhere.PostID.EQ(opts.PostID))
		}
		if opts.CommentID != "" {
			mods = append(mods, dbmodels.CommentWhere.CommentID.EQ(null.StringFrom(opts.CommentID)))
		}
	}

	mods = append(mods, qm.OrderBy(dbmodels.CommentColumns.CreatedAt+" DESC"))

	result, err := dbmodels.Comments(mods...).All(ctx, r.db)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return result, nil
}

func (r *repository) Delete(ctx context.Context, tx *sql.Tx, id string) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	c := &dbmodels.Comment{ID: id}
	_, err := c.Delete(ctx, exec)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	return nil
}

func (r *repository) GetVote(ctx context.Context, userID, commentID string) (*dbmodels.CommentVote, error) {
	var v dbmodels.CommentVote
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, comment_id, vote_type, created_at FROM comment_votes WHERE user_id = $1 AND comment_id = $2`,
		userID, commentID,
	).Scan(&v.UserID, &v.CommentID, &v.VoteType, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.CommentVotes)
	}
	return &v, nil
}

func (r *repository) AddVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx,
		`INSERT INTO comment_votes (user_id, comment_id, vote_type) VALUES ($1, $2, $3)`,
		entity.UserID, entity.CommentID, entity.VoteType,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.CommentVotes)
	}
	return nil
}

func (r *repository) UpdateVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.CommentVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx,
		`UPDATE comment_votes SET vote_type = $1 WHERE user_id = $2 AND comment_id = $3`,
		entity.VoteType, entity.UserID, entity.CommentID,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.CommentVotes)
	}
	return nil
}

func (r *repository) UpdateCounts(ctx context.Context, tx *sql.Tx, id string, upvoteDelta, downvoteDelta, expectedUpvote, expectedDownvote int) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	result, err := exec.ExecContext(ctx,
		`UPDATE comments SET upvote = upvote + $1, downvote = downvote + $2 WHERE id = $3 AND upvote = $4 AND downvote = $5`,
		upvoteDelta, downvoteDelta, id, expectedUpvote, expectedDownvote,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Comments)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domainComment.ErrConflict
	}
	return nil
}
