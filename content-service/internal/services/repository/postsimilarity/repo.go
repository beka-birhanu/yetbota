package postsimilarity

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainPostSim "github.com/beka-birhanu/yetbota/content-service/internal/domain/postsimilarity"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type repository struct {
	driver neo4j.DriverWithContext
}

type Config struct {
	Driver neo4j.DriverWithContext `validate:"required"`
}

func (c *Config) Validate() error {
	return validator.Validate.Struct(c)
}

func NewRepo(c *Config) (domainPostSim.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{driver: c.Driver}, nil
}

func (r *repository) SimilarPostsTree(ctx context.Context, postID string, maxDepth int) ([]domainPostSim.PostWithDepth, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx,
		fmt.Sprintf(`MATCH path = (p:Post {id: $postID})-[:SIMILAR_TO*1..%d]-(similar:Post)
WHERE similar.id <> $postID
RETURN similar.id AS postID, min(length(path)) AS depth`, maxDepth),
		map[string]any{"postID": postID},
	)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("postsimilarity repo: similarity tree query failed: %v", err),
		}
	}

	var posts []domainPostSim.PostWithDepth
	for result.Next(ctx) {
		rec := result.Record()
		simID, ok := rec.Get("postID")
		if !ok {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    "postsimilarity repo: similarity tree: missing postID in record",
			}
		}
		depth, ok := rec.Get("depth")
		if !ok {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    "postsimilarity repo: similarity tree: missing depth in record",
			}
		}
		posts = append(posts, domainPostSim.PostWithDepth{
			PostID: simID.(string),
			Depth:  int(depth.(int64)),
		})
	}
	if err := result.Err(); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("postsimilarity repo: similarity tree iteration failed: %v", err),
		}
	}
	return posts, nil
}
