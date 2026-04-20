package follow

import (
	"context"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	domainFollow "github.com/beka-birhanu/yetbota/identity-service/internal/domain/follow"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func paginationParams(p *domainFollow.Pagination) (skip, limit int) {
	if p == nil || p.Limit <= 0 || p.Page <= 0 {
		return 0, 100
	}
	return (p.Page - 1) * p.Limit, p.Limit
}

func collectIDs(ctx context.Context, res neo4j.ResultWithContext) ([]string, error) {
	var ids []string
	for res.Next(ctx) {
		id, _ := res.Record().Get("id")
		ids = append(ids, id.(string))
	}
	return ids, res.Err()
}

func fromNeo4jError(err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.ServerError,
		ServiceStatusCode: status.ServerErrorDatabase,
		PublicMessage:     "database error",
		ServiceMessage:    err.Error(),
	}
}
