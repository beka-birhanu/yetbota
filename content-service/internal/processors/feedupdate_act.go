package processors

import (
	"context"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainFeed "github.com/beka-birhanu/yetbota/content-service/internal/domain/feed"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
	domainPost "github.com/beka-birhanu/yetbota/content-service/internal/domain/post"
	domainPostSim "github.com/beka-birhanu/yetbota/content-service/internal/domain/postsimilarity"
	domainPostvote "github.com/beka-birhanu/yetbota/content-service/internal/domain/postvote"
	domainStorage "github.com/beka-birhanu/yetbota/content-service/internal/domain/storage"
)

type feedUpdateActivity struct {
	followerRepo  domainFollower.Repository
	postSimRepo   domainPostSim.Repository
	feedRepo      domainFeed.Repository
	postRepo      domainPost.Repository
	postvoteRepo  domainPostvote.Repository
	batchStore    domainStorage.Set
	seedBonus     float64
	qScale        float64
	epoch         int64
	halfLifeHours float64
	minFeedScore  float64
}

type feedUpdateActConfig struct {
	FollowerRepo  domainFollower.Repository `validate:"required"`
	PostSimRepo   domainPostSim.Repository  `validate:"required"`
	FeedRepo      domainFeed.Repository     `validate:"required"`
	PostRepo      domainPost.Repository     `validate:"required"`
	PostvoteRepo  domainPostvote.Repository `validate:"required"`
	BatchStore    domainStorage.Set         `validate:"required"`
	SeedBonus     float64                   `validate:"required"`
	QScale        float64                   `validate:"required"`
	Epoch         int64                     `validate:"required"`
	HalfLifeHours float64                   `validate:"required"`
	MinFeedScore  float64                   `validate:"required"`
}

func (c *feedUpdateActConfig) validate() error {
	return validator.Validate.Struct(c)
}

func newFeedUpdateAct(cfg *feedUpdateActConfig) (*feedUpdateActivity, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &feedUpdateActivity{
		followerRepo:  cfg.FollowerRepo,
		postSimRepo:   cfg.PostSimRepo,
		feedRepo:      cfg.FeedRepo,
		postRepo:      cfg.PostRepo,
		postvoteRepo:  cfg.PostvoteRepo,
		batchStore:    cfg.BatchStore,
		seedBonus:     cfg.SeedBonus,
		qScale:        cfg.QScale,
		epoch:         cfg.Epoch,
		halfLifeHours: cfg.HalfLifeHours,
		minFeedScore:  cfg.MinFeedScore,
	}, nil
}

// FetchPostFanOutData computes postScore = seedBonus + Q(q) + F(t) and returns author ID.
// Q(q) = log₂(max(q·QScale, 1)), F(t) = (createdAt − epoch) / (halfLifeHours · 3600).
func (a *feedUpdateActivity) FetchPostFanOutData(ctx context.Context, postID string) (*PostFanOutData, error) {
	post, err := a.postRepo.Read(ctx, postID)
	if err != nil {
		return nil, err
	}
	return &PostFanOutData{
		Score:    a.computePostScore(post.Likes, post.Dislikes, post.CreatedAt.Unix()),
		AuthorID: post.UserID,
	}, nil
}

// GetFollowerTree fetches the transitive follower tree for authorID, computes per-user fan-out
// scores, writes them as Redis hash batches (TTL = FanOutBatchTTLSeconds), and returns the
// batch keys. The workflow passes each key to FanOutFeedBatch — no user data crosses the
// workflow↔activity boundary.
func (a *feedUpdateActivity) GetFollowerTree(ctx context.Context, authorID string, postScore float64) ([]string, error) {
	maxDepth := a.computeMaxDepth(postScore)
	if maxDepth == 0 {
		return nil, nil
	}
	followers, err := a.followerRepo.FollowerTree(ctx, authorID, maxDepth)
	if err != nil {
		return nil, err
	}
	return a.writeUserScoreBatches(ctx, followers, postScore)
}

// SaveFanOutRecipients reads user IDs from the Redis batch at batchKey and records them as
// recipients of postID for future score-update propagation.
func (a *feedUpdateActivity) SaveFanOutRecipients(ctx context.Context, postID, batchKey string) error {
	batch, err := a.batchStore.ReadBatch(ctx, batchKey)
	if err != nil {
		return err
	}
	if len(batch) == 0 {
		return nil
	}
	userIDs := make([]string, 0, len(batch))
	for uid := range batch {
		userIDs = append(userIDs, uid)
	}
	return a.feedRepo.AddRecipients(ctx, postID, userIDs)
}

// GetSimilarPostsTree fetches the similarity graph for postID, expands each similar post's
// interactors to their followers, computes per-user fan-out scores, and writes Redis hash
// batches. Returns batch keys for FanOutFeedBatch.
func (a *feedUpdateActivity) GetSimilarPostsTree(ctx context.Context, postID string, postScore float64) ([]string, error) {
	maxDepth := a.computeMaxDepth(postScore)
	if maxDepth == 0 {
		return nil, nil
	}
	simPosts, err := a.postSimRepo.SimilarPostsTree(ctx, postID, maxDepth)
	if err != nil {
		return nil, err
	}
	if len(simPosts) == 0 {
		return nil, nil
	}

	simPostIDs := make([]string, len(simPosts))
	simDepthByPost := make(map[string]int, len(simPosts))
	for i, sp := range simPosts {
		simPostIDs[i] = sp.PostID
		simDepthByPost[sp.PostID] = sp.Depth
	}

	interactorsMap, err := a.postvoteRepo.ListVotersByPostIDs(ctx, simPostIDs)
	if err != nil {
		return nil, err
	}

	interactorBestDepth := make(map[string]int)
	for simPostID, users := range interactorsMap {
		depth := simDepthByPost[simPostID]
		for _, userID := range users {
			if existing, ok := interactorBestDepth[userID]; !ok || depth < existing {
				interactorBestDepth[userID] = depth
			}
		}
	}

	uniqueInteractors := make([]string, 0, len(interactorBestDepth))
	for userID := range interactorBestDepth {
		uniqueInteractors = append(uniqueInteractors, userID)
	}
	if len(uniqueInteractors) == 0 {
		return nil, nil
	}

	followersMap, err := a.followerRepo.FollowersOf(ctx, uniqueInteractors)
	if err != nil {
		return nil, err
	}

	// Keep max score per user across all interactor paths.
	userScores := make(map[string]float64)
	for interactorID, followers := range followersMap {
		score := postScore + distanceAttenuation(interactorBestDepth[interactorID])
		for _, followerID := range followers {
			if existing, ok := userScores[followerID]; !ok || score > existing {
				userScores[followerID] = score
			}
		}
	}

	return a.writePrecomputedScoreBatches(ctx, userScores)
}

// FanOutFeedBatch reads the user→score hash at batchKey, fans out postID to each user's feed
// via ZADD GT, and deletes the key. The key is consumed exactly once.
func (a *feedUpdateActivity) FanOutFeedBatch(ctx context.Context, postID string, batchKey string) error {
	userScores, err := a.batchStore.ReadBatch(ctx, batchKey)
	if err != nil {
		return err
	}
	if len(userScores) == 0 {
		return nil
	}
	return a.feedRepo.FanOutGT(ctx, postID, userScores)
}

