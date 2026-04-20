package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/friendsofgo/errors"
	"github.com/go-kit/kit/endpoint"

	domainAuth "github.com/beka-birhanu/yetbota/content-service/internal/domain/auth"
	ctxRP "github.com/beka-birhanu/yetbota/content-service/internal/domain/context"
)

func httpError(w http.ResponseWriter, statusCode int, err string) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(err)
}

func httpSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func AuthMiddleware(sessionManager domainAuth.SessionManager) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			data := ctx.Value(ctxRP.AppSession)
			ctxSess := data.(*ctxRP.Context)
			header := ctxSess.Header.(http.Header)["Authorization"]
			var authHeader string
			if len(header) > 0 {
				authHeader = header[0]
			}
			bearerToken := strings.Split(authHeader, " ")

			if len(bearerToken) == 2 {
				userSession, errExtract := sessionManager.ExtractUserSession(ctx, &domainAuth.TokenInfo{
					TokenType: domainAuth.AccessToken,
					Token:     authHeader,
				})
				if errExtract != nil {
					return nil, errExtract
				}

				ctxSess.UserSession = *userSession

				return next(ctx, request)
			}

			return nil, errors.New("Invalid token")
		}
	}
}

func TokenVerify(sessionManager domainAuth.SessionManager) func(context.Context, http.Handler) http.Handler {
	return func(ctx context.Context, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			bearerToken := strings.Split(authHeader, " ")

			if len(bearerToken) == 2 {
				userSession, err := sessionManager.ExtractUserSession(r.Context(), &domainAuth.TokenInfo{
					TokenType: domainAuth.AccessToken,
					Token:     authHeader,
				})
				if err != nil {
					httpError(w, http.StatusUnauthorized, err.Error())
					return
				}

				_ = userSession
				next.ServeHTTP(w, r)
			} else {
				httpError(w, http.StatusUnauthorized, "Invalid token")
			}
		})
	}
}
