package auth

import (
	"context"
	"encoding/json"
	"net/http"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	ctxYB "github.com/beka-birhanu/yetbota/identity-service/internal/domain/context"
	authSvc "github.com/beka-birhanu/yetbota/identity-service/internal/services/usecase/auth"
)

func decodeLoginHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Site     string `json:"site"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.LoginRequest{
		Username: in.Username,
		Password: in.Password,
		Site:     in.Site,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeRefreshHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
		Username     string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.RefreshRequest{
		RefreshToken: in.RefreshToken,
		Username:     in.Username,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeLogoutHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		RefreshToken string `json:"refresh_token"`
		Username     string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.LogoutRequest{
		RefreshToken: in.RefreshToken,
		Username:     in.Username,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeGenerateMobileOTPHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Mobile string `json:"mobile"`
		Random string `json:"random"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.GenerateMobileOTPRequest{
		Mobile: in.Mobile,
		Random: in.Random,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeValidateOTPHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Otp    string `json:"otp"`
		Mobile string `json:"mobile"`
		Random string `json:"random"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.ValidateOTPRequest{
		Otp:    in.Otp,
		Mobile: in.Mobile,
		Random: in.Random,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeNewPasswordHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Password string `json:"password"`
		Random   string `json:"random"`
		Mobile   string `json:"mobile"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.NewPasswordRequest{
		Password: in.Password,
		Random:   in.Random,
		Mobile:   in.Mobile,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeAuthorizationHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		Resource string `json:"resource"`
		Action   string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.AuthorizationRequest{
		Resource: in.Resource,
		Action:   in.Action,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeChangePasswordHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.ChangePasswordRequest{
		CurrentPassword: in.CurrentPassword,
		NewPassword:     in.NewPassword,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func decodeChangeMobileHTTP(ctx context.Context, r *http.Request) (any, error) {
	var in struct {
		NewMobile string `json:"new_mobile"`
		Random    string `json:"random"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		return nil, badRequest("invalid json", err)
	}
	req := &authSvc.ChangeMobileRequest{
		NewMobile: in.NewMobile,
		Random:    in.Random,
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	setCtxRequest(ctx, req)
	return req, nil
}

func setCtxRequest(ctx context.Context, req any) {
	data := ctx.Value(ctxYB.AppSession)
	ctxSess, ok := data.(*ctxYB.Context)
	if !ok || ctxSess == nil {
		return
	}
	ctxSess.SetRequest(req)
}

func badRequest(publicMsg string, err error) error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     publicMsg,
		ServiceMessage:    err.Error(),
	}
}

