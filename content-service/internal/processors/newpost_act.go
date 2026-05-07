package processors

import "github.com/beka-birhanu/yetbota/content-service/drivers/validator"

type newPostActivity struct{}

type newPostActConfig struct{}

func (c *newPostActConfig) validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func newNewPostActivity(cfg *newPostActConfig) (*newPostActivity, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &newPostActivity{}, nil
}
