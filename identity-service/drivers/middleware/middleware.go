package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/friendsofgo/errors"
	"github.com/go-kit/kit/endpoint"

	jwtLib "github.com/beka-birhanu/yetbota/identity-service/drivers/jwt"
	ctxRP "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
)

const SessionID = "Session_Id"

func httpError(w http.ResponseWriter, statusCode int, err string) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err)
}

func httpSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func AuthMiddleware() endpoint.Middleware {
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
				userSession, errExtract := jwtLib.NewExtractTokenMetadata(authHeader)
				if errExtract != nil {
					return nil, errExtract
				}

				ctxSess.UserSession = userSession

				return next(ctx, request)
			}

			return nil, errors.New("Invalid token")
		}
	}
}

func TokenVerify(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		bearerToken := strings.Split(authHeader, " ")

		if len(bearerToken) == 2 {
			// Parsing the access token metadata
			token, err := jwtLib.ParseAccessToken(authHeader)
			if err != nil {
				httpError(w, http.StatusUnauthorized, err.Error())
				return
			}

			if token.Valid {
				accessUuid, emailExtract, errExtract := jwtLib.ExtractTokenMetadata(authHeader)
				if errExtract != nil {
					httpError(w, http.StatusUnauthorized, errExtract.Error())
					return
				}

				emailAuth, errAuth := jwtLib.FetchAuth(ctx, accessUuid)
				if errAuth != nil {
					httpError(w, http.StatusUnauthorized, errAuth.Error())
					return
				}

				if emailExtract == emailAuth {
					next.ServeHTTP(w, r)
				} else {
					httpError(w, http.StatusUnauthorized, "Invalid Authentication")
					return
				}
			} else {
				httpError(w, http.StatusUnauthorized, err.Error())
				return
			}
		} else {
			httpError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
	})
}
