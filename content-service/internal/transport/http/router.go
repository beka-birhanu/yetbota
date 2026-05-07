package http

import (
	"net/http"

	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	"github.com/beka-birhanu/yetbota/content-service/internal/services/endpoint"
	httpComment "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/comment"
	httpFeed "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/feed"
	httpPost "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/post"
	httpShared "github.com/beka-birhanu/yetbota/content-service/internal/transport/http/shared"
)

type Config struct {
	BasePath       string
	E              *endpoint.Endpoints
	SessionManager domainAuth.SessionManager
	CorsHosts      []string
}

func NewRouter(cfg *Config) (http.Handler, error) {
	r := http.NewServeMux()

	v1 := http.NewServeMux()
	if cfg != nil && cfg.E != nil {
		postHandler, err := httpPost.NewHandler(&httpPost.Config{E: cfg.E, SessionManager: cfg.SessionManager})
		if err != nil {
			return nil, err
		}
		v1.Handle("/posts/", http.StripPrefix("/posts", postHandler))

		commentHandler, err := httpComment.NewHandler(&httpComment.Config{E: cfg.E, SessionManager: cfg.SessionManager})
		if err != nil {
			return nil, err
		}
		v1.Handle("/comments/", http.StripPrefix("/comments", commentHandler))

		feedHandler, err := httpFeed.NewHandler(&httpFeed.Config{E: cfg.E, SessionManager: cfg.SessionManager})
		if err != nil {
			return nil, err
		}
		v1.Handle("/feed/", http.StripPrefix("/feed", feedHandler))
	}

	if cfg != nil && cfg.BasePath != "" {
		r.Handle(cfg.BasePath+"/v1/", http.StripPrefix(cfg.BasePath+"/v1", v1))
	} else {
		r.Handle("/v1/", http.StripPrefix("/v1", v1))
	}

	var out http.Handler = r
	if cfg != nil && len(cfg.CorsHosts) > 0 {
		out = httpShared.CORS(cfg.CorsHosts)(out)
	}

	return out, nil
}
