package processors

import (
	"time"

	processorsDomain "github.com/beka-birhanu/yetbota/content-service/internal/domain/processors"
	"go.temporal.io/sdk/workflow"
)

func FeedUpdateWorkflow(ctx workflow.Context, input processorsDomain.FeedUpdateWorkflowInput) error {
	if input.TriggerEventType != string(processorsDomain.NewPostFeedUpdateTriggerEventType) {
		return nil
	}

	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    2 * time.Minute,
		ScheduleToCloseTimeout: 15 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Step 1: Fetch post score and author ID.
	var fanOutData PostFanOutData
	if err := workflow.ExecuteActivity(ctx, (*feedUpdateActivity).FetchPostFanOutData, input.PostID).Get(ctx, &fanOutData); err != nil {
		return err
	}

	logger.Info("FeedUpdateWorkflow started", "postID", input.PostID, "postScore", fanOutData.Score)

	// Step 2: Write user→score batches to Redis for both paths in parallel.
	// Each activity returns Redis batch keys; large user data never crosses this boundary.
	followerFuture := workflow.ExecuteActivity(ctx, (*feedUpdateActivity).GetFollowerTree, fanOutData.AuthorID, fanOutData.Score)
	simFuture := workflow.ExecuteActivity(ctx, (*feedUpdateActivity).GetSimilarPostsTree, input.PostID, fanOutData.Score)

	var followerKeys []string
	if err := followerFuture.Get(ctx, &followerKeys); err != nil {
		return err
	}

	var simKeys []string
	if err := simFuture.Get(ctx, &simKeys); err != nil {
		return err
	}

	// Step 3: Fan out postID to each batch in parallel.
	// FanOutFeedBatch reads the Redis hash at batchKey and runs ZADD GT.
	allKeys := append(followerKeys, simKeys...)
	fanOutFutures := make([]workflow.Future, 0, len(allKeys))
	for _, key := range allKeys {
		fanOutFutures = append(fanOutFutures, workflow.ExecuteActivity(ctx, (*feedUpdateActivity).FanOutFeedBatch, input.PostID, key))
	}
	for _, f := range fanOutFutures {
		if err := f.Get(ctx, nil); err != nil {
			logger.Error("FanOutFeedBatch failed", "error", err)
		}
	}

	// Step 4: Record recipients per batch in parallel.
	// SaveFanOutRecipients reads user IDs from the same Redis batch key (still alive via TTL).
	recipientFutures := make([]workflow.Future, 0, len(allKeys))
	for _, key := range allKeys {
		recipientFutures = append(recipientFutures, workflow.ExecuteActivity(ctx, (*feedUpdateActivity).SaveFanOutRecipients, input.PostID, key))
	}
	for _, f := range recipientFutures {
		if err := f.Get(ctx, nil); err != nil {
			logger.Error("SaveFanOutRecipients failed", "error", err)
		}
	}

	logger.Info("FeedUpdateWorkflow completed", "postID", input.PostID,
		"followerBatches", len(followerKeys), "simBatches", len(simKeys))
	return nil
}
