package follow

import (
	"context"
	"fmt"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	domainFollow "github.com/beka-birhanu/yetbota/identity-service/internal/domain/follow"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type repo struct {
	driver neo4j.DriverWithContext
}

type Config struct {
	Driver neo4j.DriverWithContext `validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

func NewRepo(c *Config) (domainFollow.Repository, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &repo{driver: c.Driver}, nil
}

func (r *repo) Follow(ctx context.Context, followerID, followeeID string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MERGE (a:User {id: $followerID})
			MERGE (b:User {id: $followeeID})
			MERGE (a)-[:FOLLOWS]->(b)
		`
		_, err := tx.Run(ctx, query, map[string]any{
			"followerID": followerID,
			"followeeID": followeeID,
		})
		return nil, err
	})
	if err != nil {
		return fromNeo4jError(err)
	}
	return nil
}

func (r *repo) Unfollow(ctx context.Context, followerID, followeeID string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (a:User {id: $followerID})-[r:FOLLOWS]->(b:User {id: $followeeID})
			DELETE r
		`
		res, err := tx.Run(ctx, query, map[string]any{
			"followerID": followerID,
			"followeeID": followeeID,
		})
		if err != nil {
			return nil, err
		}
		return res.Consume(ctx)
	})
	if err != nil {
		return fromNeo4jError(err)
	}

	if result.(neo4j.ResultSummary).Counters().RelationshipsDeleted() == 0 {
		return &toddlerr.Error{
			PublicStatusCode:  status.NotFound,
			ServiceStatusCode: status.NotFound,
			PublicMessage:     "Follow relationship not found",
			ServiceMessage:    fmt.Sprintf("no FOLLOWS edge from %s to %s", followerID, followeeID),
		}
	}
	return nil
}

func (r *repo) IsFollowing(ctx context.Context, followerID, followeeID string) (bool, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			RETURN EXISTS {
				MATCH (:User {id: $followerID})-[:FOLLOWS]->(:User {id: $followeeID})
			} AS isFollowing
		`
		res, err := tx.Run(ctx, query, map[string]any{
			"followerID": followerID,
			"followeeID": followeeID,
		})
		if err != nil {
			return false, err
		}
		if res.Next(ctx) {
			return res.Record().Values[0].(bool), nil
		}
		return false, res.Err()
	})
	if err != nil {
		return false, fromNeo4jError(err)
	}
	return result.(bool), nil
}

func (r *repo) Followers(ctx context.Context, userID string, pagination *domainFollow.Pagination) ([]string, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (f:User)-[:FOLLOWS]->(:User {id: $userID})
			RETURN f.id AS id
			ORDER BY f.id
			SKIP $skip LIMIT $limit
		`
		skip, limit := paginationParams(pagination)
		res, err := tx.Run(ctx, query, map[string]any{
			"userID": userID,
			"skip":   skip,
			"limit":  limit,
		})
		if err != nil {
			return nil, err
		}
		return collectIDs(ctx, res)
	})
	if err != nil {
		return nil, fromNeo4jError(err)
	}
	return result.([]string), nil
}

func (r *repo) Following(ctx context.Context, userID string, pagination *domainFollow.Pagination) ([]string, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (:User {id: $userID})-[:FOLLOWS]->(f:User)
			RETURN f.id AS id
			ORDER BY f.id
			SKIP $skip LIMIT $limit
		`
		skip, limit := paginationParams(pagination)
		res, err := tx.Run(ctx, query, map[string]any{
			"userID": userID,
			"skip":   skip,
			"limit":  limit,
		})
		if err != nil {
			return nil, err
		}
		return collectIDs(ctx, res)
	})
	if err != nil {
		return nil, fromNeo4jError(err)
	}
	return result.([]string), nil
}

func (r *repo) CountFollowers(ctx context.Context, userID string) (int64, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (f:User)-[:FOLLOWS]->(:User {id: $userID})
			RETURN count(f) AS cnt
		`
		res, err := tx.Run(ctx, query, map[string]any{"userID": userID})
		if err != nil {
			return int64(0), err
		}
		if res.Next(ctx) {
			return res.Record().Values[0].(int64), nil
		}
		return int64(0), res.Err()
	})
	if err != nil {
		return 0, fromNeo4jError(err)
	}
	return result.(int64), nil
}

func (r *repo) CountFollowing(ctx context.Context, userID string) (int64, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := `
			MATCH (:User {id: $userID})-[:FOLLOWS]->(f:User)
			RETURN count(f) AS cnt
		`
		res, err := tx.Run(ctx, query, map[string]any{"userID": userID})
		if err != nil {
			return int64(0), err
		}
		if res.Next(ctx) {
			return res.Record().Values[0].(int64), nil
		}
		return int64(0), res.Err()
	})
	if err != nil {
		return 0, fromNeo4jError(err)
	}
	return result.(int64), nil
}
