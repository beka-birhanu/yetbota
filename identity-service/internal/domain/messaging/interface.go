package messaging

import "context"

type SMSClient interface {
	SendOTP(ctx context.Context, mobile string, otp string) error
}
