package follower

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	domainFollower "github.com/beka-birhanu/yetbota/content-service/internal/domain/follower"
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

func NewRepo(c *Config) (domainFollower.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repository{driver: c.Driver}, nil
}

func (r *repository) FollowerTree(ctx context.Context, authorID string, maxDepth int) ([]domainFollower.UserWithDepth, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx,
		fmt.Sprintf(`MATCH path = (follower:User)-[:FOLLOWS*1..%d]->(author:User {id: $authorID})
RETURN follower.id AS userID, min(length(path)) AS depth`, maxDepth),
		map[string]any{"authorID": authorID},
	)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("follower repo: follower tree query failed: %v", err),
		}
	}

	var users []domainFollower.UserWithDepth
	for result.Next(ctx) {
		rec := result.Record()
		userID, ok := rec.Get("userID")
		if !ok {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    "follower repo: follower tree: missing userID in record",
			}
		}
		depth, ok := rec.Get("depth")
		if !ok {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    "follower repo: follower tree: missing depth in record",
			}
		}
		users = append(users, domainFollower.UserWithDepth{
			UserID: userID.(string),
			Depth:  int(depth.(int64)),
		})
	}
	if err := result.Err(); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("follower repo: follower tree iteration failed: %v", err),
		}
	}
	return users, nil
}

func (r *repository) FollowersOf(ctx context.Context, userIDs []string) (map[string][]string, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer func() { _ = session.Close(ctx) }()

	result, err := session.Run(ctx,
		`MATCH (follower:User)-[:FOLLOWS]->(u:User)
WHERE u.id IN $userIDs
RETURN u.id AS userID, follower.id AS followerID`,
		map[string]any{"userIDs": userIDs},
	)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("follower repo: followers of users query failed: %v", err),
		}
	}

	res := make(map[string][]string)
	for result.Next(ctx) {
		rec := result.Record()
		userID, ok := rec.Get("userID")
		if !ok {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    "follower repo: followers of users: missing userID in record",
			}
		}
		followerID, ok := rec.Get("followerID")
		if !ok {
			return nil, &toddlerr.Error{
				PublicStatusCode:  status.ServerError,
				ServiceStatusCode: status.ServerError,
				PublicMessage:     "something went wrong",
				ServiceMessage:    "follower repo: followers of users: missing followerID in record",
			}
		}
		res[userID.(string)] = append(res[userID.(string)], followerID.(string))
	}
	if err := result.Err(); err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "something went wrong",
			ServiceMessage:    fmt.Sprintf("follower repo: followers of users iteration failed: %v", err),
		}
	}
	return res, nil
}
