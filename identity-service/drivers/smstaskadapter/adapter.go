package smstaskadapter

import (
	"context"
	"fmt"
)

type Adapeter struct{}

type Config struct{}

func NewAdapter(c *Config) *Adapeter {
	return &Adapeter{}
}

func (a *Adapeter) SendOTP(ctx context.Context, mobile string, otp string) error {
	fmt.Printf("Sending %s to mobile: %s\n", otp, mobile)
	return nil
}
