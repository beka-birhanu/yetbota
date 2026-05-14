package feed

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
		getEndpoint := driverMW.AuthMiddleware(cfg.SessionManager)(cfg.E.FeedGet)
		mux.Handle("GET /", kithttp.NewServer(
			getEndpoint,
			decodeFeedGetHTTP,
			encodeFeedGetHTTP,
			shared.ServerOptions()...,
		))

		markViewedEndpoint := driverMW.AuthMiddleware(cfg.SessionManager)(cfg.E.FeedMarkViewed)
		mux.Handle("POST /viewed", kithttp.NewServer(
			markViewedEndpoint,
			decodeFeedMarkViewedHTTP,
			encodeFeedMarkViewedHTTP,
			shared.ServerOptions()...,
		))
	}

	return mux, nil
}
