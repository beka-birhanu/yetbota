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

		if cfg.SessionManager != nil {
			addEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(addEndpoint)
			readEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(readEndpoint)
			listEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(listEndpoint)
			deleteEndpoint = driverMW.AuthMiddleware(cfg.SessionManager)(deleteEndpoint)
		}

		mux.Handle("POST /", kithttp.NewServer(addEndpoint, decodeCommentAddHTTP, encodeCommentAddHTTP, shared.ServerOptions()...))
		mux.Handle("GET /", kithttp.NewServer(listEndpoint, decodeCommentListHTTP, encodeCommentListHTTP, shared.ServerOptions()...))
		mux.Handle("GET /{id}", kithttp.NewServer(readEndpoint, decodeCommentReadHTTP, encodeCommentReadHTTP, shared.ServerOptions()...))
		mux.Handle("DELETE /{id}", kithttp.NewServer(deleteEndpoint, decodeCommentDeleteHTTP, encodeCommentDeleteHTTP, shared.ServerOptions()...))
	}

	return mux, nil
}
