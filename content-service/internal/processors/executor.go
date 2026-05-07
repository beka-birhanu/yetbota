package processors

import (
	"context"
	"fmt"
	"net/http"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
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
	Client client.Client `validate:"required"`
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
	newPostAct, err := newNewPostActivity(&newPostActConfig{})
	if err != nil {
		return nil, err
	}
	feedUpdateAct, err := newFeedUpdateAct(&feedUpdateActConfig{})
	if err != nil {
		return nil, err
	}
	return &Executor{
		client:             cfg.Client,
		newPostActivity:    newPostAct,
		feedUpdateActivity: feedUpdateAct,
	}, nil
}

// RegisterWorkflowsAndActivity registers the workflow and all activities on the worker.
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
			ServiceMessage:    fmt.Sprintf("Error starting order workflow: %v", err),
		}
	}

	return nil
}

// TriggerNewPostWorkflow implements [processors.Executor].
func (e *Executor) TriggerNewPostWorkflow(ctx context.Context, input processors.NewPostWorkflowInput) error {
	workflowOptions := client.StartWorkflowOptions{
		ID:                    fmt.Sprintf("NEW-POST-%s-%s", input.PostID, time.Now().Format(time.RFC3339)),
		TaskQueue:             constants.FeedUpdateWorkflowQueue,
		WorkflowTaskTimeout:   10 * time.Second,
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}

	_, err := e.client.ExecuteWorkflow(ctx, workflowOptions, NewPostWorkflow, input)
	if err != nil {
		return &toddlerr.Error{
			PublicStatusCode:  http.StatusInternalServerError,
			ServiceStatusCode: http.StatusInternalServerError,
			PublicMessage:     "Something went wrong.",
			ServiceMessage:    fmt.Sprintf("Error starting order workflow: %v", err),
		}
	}
	return nil
}
