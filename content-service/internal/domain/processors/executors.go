package processors

import "context"

type FeedUpdateTriggerEventType string

const (
	NewPostFeedUpdateTriggerEventType     FeedUpdateTriggerEventType = "NEW_POST"
	InteractionFeedUpdateTriggerEventType FeedUpdateTriggerEventType = "INTERACTION"
)

type NewPostWorkflowInput struct {
	PostID string
}

type FeedUpdateWorkflowInput struct {
	PostID           string
	InteractorID     string
	TriggerEventType string
}

type PostEmbeddingWorkflowInput struct {
	PostID string
}

type Executor interface {
	TriggerNewPostWorkflow(ctx context.Context, input NewPostWorkflowInput) error
	TriggerFeedUpdateWorkflow(ctx context.Context, input FeedUpdateWorkflowInput) error
}
