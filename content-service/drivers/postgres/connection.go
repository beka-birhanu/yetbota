package postgres

import (
	"database/sql"
	"fmt"

	"github.com/beka-birhanu/yetbota/content-service/drivers/validator"
	_ "github.com/lib/pq"
)

const postgresDBMS = "postgres"

type Config struct {
	Host     string `yaml:"host" mapstructure:"host" validate:"required"`
	Port     string `yaml:"port" mapstructure:"port" validate:"required"`
	User     string `yaml:"user" mapstructure:"user" validate:"required"`
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
	DB       string `yaml:"db" mapstructure:"db" validate:"required"`
}

func (c *Config) Validate() error {
	if err := validator.Validate.Struct(c); err != nil {
		return err
	}
	return nil
}

func (c *Config) getDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DB)
}

func NewDB(c *Config) (*sql.DB, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	db, err := sql.Open(postgresDBMS, c.getDSN())
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
