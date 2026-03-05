package auth

import (
	"context"
	"time"

	"github.com/aarondl/null/v8"
)

type Otp struct {
	SessionID      string    `json:"session_id"`
	Otp            string    `json:"otp"`
	Random         string    `json:"random"`
	GeneratedCount int32     `json:"generated_count"`
	ErrorCount     int32     `json:"error_count"`
	LockedUntil    null.Time `json:"locked_until"`
	IssuedAt       time.Time `json:"issued_at"`
	Verified       bool      `json:"verified"`
}

type OtpStore interface {
	// Read returns the OTP record for the given identifier.
	// returns zero, nil if not found.
	Read(ctx context.Context, key string) (Otp, error)
	Save(ctx context.Context, otp Otp, key string, ttl time.Duration) error
}
