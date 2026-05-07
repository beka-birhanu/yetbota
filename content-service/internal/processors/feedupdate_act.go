package processors

import "github.com/beka-birhanu/yetbota/content-service/drivers/validator"

type feedUpdateActivity struct{}

type feedUpdateActConfig struct{}

func (c *feedUpdateActConfig) validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func newFeedUpdateAct(cfg *feedUpdateActConfig) (*feedUpdateActivity, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &feedUpdateActivity{}, nil
}
