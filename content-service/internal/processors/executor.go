package processors

import (
	"context"
	"fmt"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainFeed "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
	domainPhoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/photo"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostphoto "github.com/beka-birhanu/yetbota/content-service/internal/domain/postphoto"
	domainPostSim "github.com/beka-birhanu/yetbota/content-service/internal/domain/postsimilarity"
	domainPostvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type Executor struct {
	client             client.Client
	newPostActivity    *newPostActivity
	feedUpdateActivity *feedUpdateActivity
}

type Config struct {
	Client        client.Client              `validate:"required"`
	PostPhotoRepo domainPostphoto.Repository `validate:"required"`
	PhotoRepo     domainPhoto.Repository     `validate:"required"`
	Bucket        domainStorage.Bucket       `validate:"required"`
	BucketName    string                     `validate:"required"`
	BucketRegion  string                     `validate:"required"`
	FollowerRepo  domainFollower.Repository  `validate:"required"`
	PostSimRepo   domainPostSim.Repository   `validate:"required"`
	FeedRepo      domainFeed.Repository      `validate:"required"`
	PostRepo      domainPost.Repository      `validate:"required"`
	PostvoteRepo  domainPostvote.Repository  `validate:"required"`
	BatchStore    domainStorage.Set          `validate:"required"`
	SeedBonus     float64                    `validate:"required"`
	QScale        float64                    `validate:"required"`
	Epoch         int64                      `validate:"required"`
	HalfLifeHours float64                    `validate:"required"`
	MinFeedScore  float64                    `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func NewExecutor(cfg *Config) (*Executor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	newPostAct, err := newNewPostActivity(&newPostActConfig{
		PostPhotoRepo: cfg.PostPhotoRepo,
		PhotoRepo:     cfg.PhotoRepo,
		Bucket:        cfg.Bucket,
		BucketName:    cfg.BucketName,
		BucketRegion:  cfg.BucketRegion,
	})
	if err != nil {
		return nil, err
	}
	feedUpdateAct, err := newFeedUpdateAct(&feedUpdateActConfig{
		FollowerRepo:  cfg.FollowerRepo,
		PostSimRepo:   cfg.PostSimRepo,
		FeedRepo:      cfg.FeedRepo,
		PostRepo:      cfg.PostRepo,
		PostvoteRepo:  cfg.PostvoteRepo,
		BatchStore:    cfg.BatchStore,
		SeedBonus:     cfg.SeedBonus,
		QScale:        cfg.QScale,
		Epoch:         cfg.Epoch,
		HalfLifeHours: cfg.HalfLifeHours,
		MinFeedScore:  cfg.MinFeedScore,
	})
	if err != nil {
		return nil, err
	}
	return &Executor{
		client:             cfg.Client,
		newPostActivity:    newPostAct,
		feedUpdateActivity: feedUpdateAct,
	}, nil
}

// RegisterWorkflowsAndActivity registers all workflows and activities on the worker.
func (e *Executor) RegisterWorkflowsAndActivity(w worker.Worker) {
	w.RegisterActivity(e.newPostActivity)
	w.RegisterActivity(e.feedUpdateActivity)
	w.RegisterWorkflow(NewPostWorkflow)
	w.RegisterWorkflow(FeedUpdateWorkflow)
}

// TriggerFeedUpdateWorkflow implements [processors.Executor].
func (e *Executor) TriggerFeedUpdateWorkflow(ctx context.Context, input processors.FeedUpdateWorkflowInput) error {
	workflowOptions := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("FEED-UPDATE-%s-%s-%s", input.PostID, input.InteractorID, time.Now().Format(time.RFC3339)),
		TaskQueue:             constants.FeedUpdateWorkflowQueue,
		WorkflowTaskTimeout:   10 * time.Second,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	_, err := e.client.ExecuteWorkflow(ctx, workflowOptions, FeedUpdateWorkflow, input)
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  http.StatusInternalServerError,
			ServiceStatusCode: http.StatusInternalServerError,
			PublicMessage:     "Something went wrong.",
			ServiceMessage:    fmt.Sprintf("Error starting feed update workflow: %v", err),
		}
	}

	return nil
}

// TriggerNewPostWorkflow implements [processors.Executor].
func (e *Executor) TriggerNewPostWorkflow(ctx context.Context, input processors.NewPostWorkflowInput) error {
	workflowOptions := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("NEW-POST-%s-%s", input.PostID, time.Now().Format(time.RFC3339)),
		TaskQueue:             constants.NewPostWorkflowQueue,
		WorkflowTaskTimeout:   10 * time.Second,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	_, err := e.client.ExecuteWorkflow(ctx, workflowOptions, NewPostWorkflow, input)
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  http.StatusInternalServerError,
			ServiceStatusCode: http.StatusInternalServerError,
			PublicMessage:     "Something went wrong.",
			ServiceMessage:    fmt.Sprintf("Error starting new post workflow: %v", err),
		}
	}
	return nil
}
