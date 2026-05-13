package processors

import (
	"time"

	"github.com/beka-birhanu/yetbota/content-service/drivers/constants"
	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
)

func NewPostWorkflow(ctx workflow.Context, input processorsDomain.NewPostWorkflowInput) error {
	logger := workflow.GetLogger(ctx)

	defaultAO := workflow.ActivityOptions{
		StartToCloseTimeout:    time.Minute,
		ScheduleToCloseTimeout: 6 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, defaultAO)

	logger.Info("NewPostWorkflow started", "postID", input.PostID)

	// Step 1: Fetch photo IDs for this post.
	photosFuture := workflow.ExecuteActivity(ctx, (*newPostActivity).FetchPostPhotoIDs, input.PostID)

	var photoIDs []string
	if err := photosFuture.Get(ctx, &photoIDs); err != nil {
		return err
	}

	// Step 2: Compress and upload mobile/web variants for each photo in parallel.
	if len(photoIDs) > 0 {
		compressCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			StartToCloseTimeout:    3 * time.Minute,
			ScheduleToCloseTimeout: 10 * time.Minute,
		})
		futures := make([]workflow.Future, len(photoIDs))
		for i, id := range photoIDs {
			futures[i] = workflow.ExecuteActivity(compressCtx, (*newPostActivity).ProcessPhoto, id)
		}
		for _, f := range futures {
			if err := f.Get(ctx, nil); err != nil {
				return err
			}
		}
	}

	// Step 3: Build embedding input and trigger AI embedding child workflow.
	// Blocking — graph links must exist before feed update runs.
	embeddingCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue: constants.PostEmbeddingWorkflowQueue,
	})
	if err := workflow.ExecuteChildWorkflow(embeddingCtx,
		constants.PostEmbeddingWorkflowName,
		processorsDomain.PostEmbeddingWorkflowInput(input)).Get(ctx, nil); err != nil {
		logger.Error("PostEmbedding workflow failed, continuing to feed update", "postID", input.PostID, "error", err)
	}

	// Step 4: Trigger feed update child workflow.
	feedInput := processorsDomain.FeedUpdateWorkflowInput{
		PostID:           input.PostID,
		TriggerEventType: string(processorsDomain.NewPostFeedUpdateTriggerEventType),
	}
	feedCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue:         constants.FeedUpdateWorkflowQueue,
		ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
	})
	if err := workflow.ExecuteChildWorkflow(feedCtx, FeedUpdateWorkflow, feedInput).Get(ctx, nil); err != nil {
		logger.Error("FeedUpdateWorkflow failed, continuing to post", "postID", input.PostID, "error", err)
		return err
	}

	logger.Info("NewPostWorkflow completed", "postID", input.PostID)
	return nil
}
