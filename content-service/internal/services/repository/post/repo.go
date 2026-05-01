package post

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/dbmodels"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	"github.com/lib/pq"
)

type repo struct {
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

func NewRepo(cfg *Config) (domainPost.Repository, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &repo{db: cfg.DB}, nil
}

func (r *repo) Add(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	if err := entity.Insert(ctx, exec, boil.Infer()); err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}
	return nil
}

func (r *repo) Read(ctx context.Context, id string) (*dbmodels.Post, error) {
	post, err := dbmodels.FindPost(ctx, r.db, id)
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}
	return post, nil
}

func (r *repo) Update(ctx context.Context, tx *sql.Tx, entity *dbmodels.Post) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}

	rowAff, err := entity.Update(ctx, exec, boil.Infer())
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	if rowAff == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Post not found",
			ServiceMessage:    fmt.Sprintf("post not found id: %s", entity.ID),
		}
	}

	return nil
}

func (r *repo) GetVote(ctx context.Context, userID, postID string) (*dbmodels.PostVote, error) {
	var v dbmodels.PostVote
	err := r.db.QueryRowContext(ctx,
		`SELECT user_id, post_id, vote_type, created_at FROM post_votes WHERE user_id = $1 AND post_id = $2`,
		userID, postID,
	).Scan(&v.UserID, &v.PostID, &v.VoteType, &v.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return &v, nil
}

func (r *repo) AddVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx,
		`INSERT INTO post_votes (user_id, post_id, vote_type) VALUES ($1, $2, $3)`,
		entity.UserID, entity.PostID, entity.VoteType,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return nil
}

func (r *repo) UpdateVote(ctx context.Context, tx *sql.Tx, entity *dbmodels.PostVote) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	_, err := exec.ExecContext(ctx,
		`UPDATE post_votes SET vote_type = $1 WHERE user_id = $2 AND post_id = $3`,
		entity.VoteType, entity.UserID, entity.PostID,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.PostVotes)
	}
	return nil
}

func (r *repo) List(ctx context.Context, opts *domainPost.ListOptions) ([]*dbmodels.Post, int64, error) {
	var filterMods []qm.QueryMod

	if opts.UserID != "" {
		filterMods = append(filterMods, dbmodels.PostWhere.UserID.EQ(opts.UserID))
	}
	if opts.Search != "" {
		pat := "%" + opts.Search + "%"
		filterMods = append(filterMods, qm.Where("(title ILIKE ? OR description ILIKE ?)", pat, pat))
	}
	if opts.IsQuestion != nil {
		filterMods = append(filterMods, dbmodels.PostWhere.IsQuestion.EQ(*opts.IsQuestion))
	}
	if len(opts.Tags) > 0 {
		filterMods = append(filterMods, qm.Where("tags && ?", pq.Array(opts.Tags)))
	}
	if opts.NearLat != nil && opts.NearLon != nil && opts.RadiusKm != nil {
		filterMods = append(filterMods, qm.Where(
			"location IS NOT NULL AND ST_DWithin(location::geography, ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography, ?)",
			*opts.NearLon, *opts.NearLat, *opts.RadiusKm*1000,
		))
	}

	total, err := dbmodels.Posts(filterMods...).Count(ctx, r.db)
	if err != nil {
		return nil, 0, toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	sortField := "created_at"
	switch opts.SortField {
	case domainPost.ListSortFieldLikes:
		sortField = "likes"
	case domainPost.ListSortFieldDislikes:
		sortField = "dislikes"
	case domainPost.ListSortFieldComments:
		sortField = "comment_count"
	}
	sortDir := "DESC"
	if opts.SortDir == domainPost.ListSortDirAsc {
		sortDir = "ASC"
	}

	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	page := opts.Page
	if page <= 0 {
		page = 1
	}

	allMods := append(filterMods,
		qm.OrderBy(fmt.Sprintf("%s %s", sortField, sortDir)),
		qm.Limit(pageSize),
		qm.Offset((page-1)*pageSize),
	)

	posts, err := dbmodels.Posts(allMods...).All(ctx, r.db)
	if err != nil {
		return nil, 0, toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}

	return posts, total, nil
}

func (r *repo) UpdateCounts(ctx context.Context, tx *sql.Tx, id string, likesDelta, dislikesDelta, expectedLikes, expectedDislikes int) error {
	var exec boil.ContextExecutor = r.db
	if tx != nil {
		exec = tx
	}
	result, err := exec.ExecContext(ctx,
		`UPDATE posts SET likes = likes + $1, dislikes = dislikes + $2 WHERE id = $3 AND likes = $4 AND dislikes = $5`,
		likesDelta, dislikesDelta, id, expectedLikes, expectedDislikes,
	)
	if err != nil {
		return toddlerr.FromDBError(err, dbmodels.TableNames.Posts)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domainPost.ErrConflict
	}
	return nil
}
