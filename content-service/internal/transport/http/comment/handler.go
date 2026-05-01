package comment

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
		addEndpoint := cfg.E.CommentAdd
		readEndpoint := cfg.E.CommentRead
		listEndpoint := cfg.E.CommentList
		deleteEndpoint := cfg.E.CommentDelete
		voteEndpoint := cfg.E.CommentVote

		if cfg.SessionManager != nil {
			addEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(addEndpoint)
			readEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(readEndpoint)
			listEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(listEndpoint)
			deleteEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(deleteEndpoint)
			voteEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(voteEndpoint)
		}

		addServer := kithttp.NewServer(
			addEndpoint,
			decodeCommentAddHTTP,
			encodeCommentAddHTTP,
			shared.ServerOptions()...,
		)
		readServer := kithttp.NewServer(
			readEndpoint,
			decodeCommentReadHTTP,
			encodeCommentReadHTTP,
			shared.ServerOptions()...,
		)
		listServer := kithttp.NewServer(
			listEndpoint,
			decodeCommentListHTTP,
			encodeCommentListHTTP,
			shared.ServerOptions()...,
		)
		deleteServer := kithttp.NewServer(
			deleteEndpoint,
			decodeCommentDeleteHTTP,
			encodeCommentDeleteHTTP,
			shared.ServerOptions()...,
		)
		voteServer := kithttp.NewServer(
			voteEndpoint,
			decodeCommentVoteHTTP,
			encodeCommentVoteHTTP,
			shared.ServerOptions()...,
		)

		mux.Handle("POST /", addServer)
		mux.Handle("GET /", listServer)
		mux.Handle("GET /{id}", readServer)
		mux.Handle("DELETE /{id}", deleteServer)
		mux.Handle("POST /{id}/vote", voteServer)
	}

	return mux, nil
}

