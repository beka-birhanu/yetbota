package auth

import (
	"context"
	"encoding/json"
	"time"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/beka-birhanu/yetbota/identity-service/internal/domain/auth"
	"github.com/go-redis/redis/v8"
)

type OtpStoreConfig struct {
	RedisConn *redis.Client `validate:"required"`
}

func (c *OtpStoreConfig) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return toddlerr.FromValidationErrors(err)
	}
	return nil
}

type otpStore struct {
	redisConn *redis.Client
}

func NewOtpStore(cfg *OtpStoreConfig) (auth.OtpStore, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &otpStore{
		redisConn: cfg.RedisConn,
	}, nil
}

func (o *otpStore) Read(ctx context.Context, key string) (auth.Otp, error) {
	val, err := o.redisConn.Get(ctx, key).Result()
	if err == redis.Nil {
		return auth.Otp{}, nil
	}
	if err != nil {
		return auth.Otp{}, err
	}

	var otp auth.Otp
	if err := json.Unmarshal([]byte(val), &otp); err != nil {
		return auth.Otp{}, err
	}

	return otp, nil
}

func (o *otpStore) Save(ctx context.Context, otp auth.Otp, key string, ttl time.Duration) error {
	data, err := json.Marshal(otp)
	if err != nil {
		return err
	}

	return o.redisConn.Set(ctx, key, data, ttl).Err()
}
