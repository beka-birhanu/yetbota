package post

import (
	"net/http"

	driverMW "github.com/beka-birhanu/yetbota/content-service/drivers/middleware"
	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	"github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
	kithttp "github.com/go-kit/kit/transport/http"
)

type Config struct {
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
}

func NewHandler(cfg *Config) (http.Handler, error) {
	mux := http.NewServeMux()

	if cfg != nil && cfg.E != nil {
		addEndpoint    := cfg.E.PostAdd
		readEndpoint   := cfg.E.PostRead
		updateEndpoint := cfg.E.PostUpdate
		voteEndpoint   := cfg.E.PostVote
		listEndpoint   := cfg.E.PostList

		if cfg.SessionManager != nil {
			// Read and List are public — no auth required (matches gRPC SkipAuthGrpc)
			addEndpoint    = driverMW.AuthMiddleware(cfg.SessionManager)(addEndpoint)
			updateEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(updateEndpoint)
			voteEndpoint   = driverMW.AuthMiddleware(cfg.SessionManager)(voteEndpoint)
		}

		opts := shared.ServerOptions()

		mux.Handle("POST /",          kithttp.NewServer(addEndpoint,    decodePostAddHTTP,    encodePostAddHTTP,    opts...))
		mux.Handle("GET /{id}",       kithttp.NewServer(readEndpoint,   decodePostReadHTTP,   encodePostReadHTTP,   opts...))
		mux.Handle("PATCH /{id}",     kithttp.NewServer(updateEndpoint, decodePostUpdateHTTP, encodePostUpdateHTTP, opts...))
		mux.Handle("POST /{id}/vote", kithttp.NewServer(voteEndpoint,   decodePostVoteHTTP,   encodePostVoteHTTP,   opts...))
		mux.Handle("GET /",           kithttp.NewServer(listEndpoint,   decodePostListHTTP,   encodePostListHTTP,   opts...))
	}

	return mux, nil
}
