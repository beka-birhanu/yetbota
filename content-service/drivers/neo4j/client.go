package neo4j

import (
	"context"
	"fmt"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Config struct {
	URI      string `yaml:"uri"      mapstructure:"uri"      validate:"required"`
	Username string `yaml:"username" mapstructure:"username" validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
}

func (c *Config) Validate() error {
	return validator.Validate.Struct(c)
}

func NewDriver(c *Config) (neo4j.DriverWithContext, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	driver, err := neo4j.NewDriverWithContext(c.URI, neo4j.BasicAuth(c.Username, c.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("neo4j driver: %w", err)
	}
	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		_ = driver.Close(context.Background())
		return nil, fmt.Errorf("neo4j connectivity: %w", err)
	}
	return driver, nil
}
