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
		addEndpoint := cfg.E.PostAdd
		readEndpoint := cfg.E.PostRead
		updateEndpoint := cfg.E.PostUpdate
		voteEndpoint := cfg.E.PostVote

		if cfg.SessionManager != nil {
			addEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(addEndpoint)
			readEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(readEndpoint)
			updateEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(updateEndpoint)
			voteEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(voteEndpoint)
		}

		addServer := kithttp.NewServer(
			addEndpoint,
			decodePostAddHTTP,
			encodePostAddHTTP,
			shared.ServerOptions()...,
		)
		readServer := kithttp.NewServer(
			readEndpoint,
			decodePostReadHTTP,
			encodePostReadHTTP,
			shared.ServerOptions()...,
		)
		updateServer := kithttp.NewServer(
			updateEndpoint,
			decodePostUpdateHTTP,
			encodePostUpdateHTTP,
			shared.ServerOptions()...,
		)
		voteServer := kithttp.NewServer(
			voteEndpoint,
			decodePostVoteHTTP,
			encodePostVoteHTTP,
			shared.ServerOptions()...,
		)

		mux.Handle("POST /", addServer)
		mux.Handle("GET /{id}", readServer)
		mux.Handle("PATCH /{id}", updateServer)
		mux.Handle("POST /{id}/vote", voteServer)
	}

	return mux, nil
}
