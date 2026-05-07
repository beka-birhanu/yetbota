package processors

import "context"

type NewPostWorkflowInput struct {
	PostID string
}

type FeedUpdateWorkflowInput struct {
	PostID          string
	InteractorID    string
	InteractionType string
}

type Executor interface {
	TriggerNewPostWorkflow(ctx context.Context, input NewPostWorkflowInput) error
	TriggerFeedUpdateWorkflow(ctx context.Context, input FeedUpdateWorkflowInput) error
}
