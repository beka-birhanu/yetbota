package neo4j

import (
	"context"
	"fmt"

	"github.com/beka-birhanu/yetbota/identity-service/drivers/validator"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Config struct {
	URI      string `yaml:"uri" mapstructure:"uri" validate:"required"`
	Username string `yaml:"username" mapstructure:"username" validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func NewDriver(c *Config) (neo4j.DriverWithContext, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	driver, err := neo4j.NewDriverWithContext(c.URI, neo4j.BasicAuth(c.Username, c.Password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create neo4j driver: %w", err)
	}

	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("failed to connect to neo4j: %w", err)
	}

	if err := applyConstraints(ctx, driver); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("failed to apply neo4j constraints: %w", err)
	}

	return driver, nil
}

func applyConstraints(ctx context.Context, driver neo4j.DriverWithContext) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		_ = session.Close(ctx)
	}()

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, `
			CREATE CONSTRAINT user_id_unique IF NOT EXISTS
			FOR (u:User)
			REQUIRE u.id IS UNIQUE
		`, nil)
		return nil, err
	})
	return err
}
