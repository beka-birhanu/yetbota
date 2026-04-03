package rest

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/beka-birhanu/yetbota/ai-service/drivers/config"
)

type Handler struct {
	cfg *config.Configs
	mux *http.ServeMux
}

func NewHandler(cfg *config.Configs) *Handler {
	h := &Handler{cfg: cfg, mux: http.NewServeMux()}

	h.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	h.mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	h.mux.HandleFunc("GET /version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"name":    cfg.App.Name,
			"version": cfg.App.Version,
		})
	})

	notImpl := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":    "99",
			"success": false,
			"message": "not implemented (phase 2+)",
		})
	}

	h.mux.HandleFunc("POST /v1/embeddings", notImpl)
	h.mux.HandleFunc("POST /v1/vector/upsert", notImpl)
	h.mux.HandleFunc("POST /v1/vector/search", notImpl)
	h.mux.HandleFunc("POST /v1/duplicate/candidates", notImpl)
	h.mux.HandleFunc("POST /v1/chat", notImpl)

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" && isAllowedOrigin(origin, h.cfg.Cors.Hosts) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.mux.ServeHTTP(w, r)
}

func isAllowedOrigin(origin string, hosts []string) bool {
	for _, h := range hosts {
		if origin == h {
			return true
		}
		if strings.HasPrefix(origin, "http://"+h) || strings.HasPrefix(origin, "https://"+h) {
			return true
		}
	}
	return false
}