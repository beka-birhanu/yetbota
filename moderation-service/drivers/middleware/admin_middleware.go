package middleware

import (
	"context"
	"net/http"

	"github.com/beka-birhanu/yetbota/moderation-service/drivers/constants"
	jwtlib "github.com/beka-birhanu/yetbota/moderation-service/drivers/jwt"
	ctxRP "github.com/beka-birhanu/yetbota/moderation-service/internal/domain/context"
	"github.com/go-kit/kit/endpoint"
)

// AdminAuthMiddlewareEndpoint returns a middleware function that checks admin authentication
func AdminAuthMiddlewareEndpoint() endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			// Retrieve context session injected in serverBefore
			data := ctx.Value(ctxRP.AppSession)
			if data == nil {
				return nil, &Error{Message: "Unauthorized", Code: http.StatusUnauthorized}
			}

			ctxSess := data.(*ctxRP.Context)

			// Extract Authorization header saved earlier
			headers, ok := ctxSess.Header.(http.Header)
			if !ok {
				return nil, &Error{Message: "Unauthorized", Code: http.StatusUnauthorized}
			}
			authHeader := headers.Get("Authorization")
			if authHeader == "" {
				return nil, &Error{Message: "Unauthorized", Code: http.StatusUnauthorized}
			}

			// Parse and validate JWT; extract user session
			userSession, err := jwtlib.ExtractAdminTokenMetadata(authHeader)
			if err != nil {
				return nil, &Error{Message: "Unauthorized", Code: http.StatusUnauthorized}
			}

			if _, err := jwtlib.FetchAuth(ctx, userSession.AccessUUID); err != nil {
				return nil, &Error{Message: "Unauthorized", Code: http.StatusUnauthorized}
			}

			// Attach to context session
			ctxSess.UserSession = userSession

			// Enforce admin role
			if ctxSess.UserSession.RoleID != constants.RoleAdmin {
				return nil, &Error{Message: "Forbidden - Admin access required", Code: http.StatusForbidden}
			}

			return next(ctx, request)
		}
	}
}

// Error type for middleware operations
type Error struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *Error) Error() string {
	return e.Message
}
