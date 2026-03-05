package main

import (
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/beka-birhanu/yetbota/moderation-service/drivers/config"
	"github.com/beka-birhanu/yetbota/moderation-service/drivers/constants"
	"github.com/beka-birhanu/yetbota/moderation-service/drivers/dbmigrations"
	"github.com/beka-birhanu/yetbota/moderation-service/drivers/logger"
	"github.com/beka-birhanu/yetbota/moderation-service/drivers/postgres"
	"github.com/beka-birhanu/yetbota/moderation-service/drivers/validator"
	"github.com/pressly/goose"
)

func main() {
	validator.InitValidator()
	logger.InitLogger()

	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("error load config: %v", err))
	}

	pgdb, err := postgres.NewDBMaster(
		&postgres.Config{
			Host:     cfg.Postgres.Host,
			Port:     cfg.Postgres.Port,
			User:     cfg.Postgres.User,
			Password: cfg.Postgres.Password,
			DB:       cfg.Postgres.DB,
		})
	if err != nil {
		panic(fmt.Errorf("error connect postgres: %v", err))
	}
	defer func() {
		_ = pgdb.Close()
	}()

	boil.SetDB(pgdb.DB)

	if err := pgdb.Ping(); err != nil {
		panic(fmt.Errorf("error pinging database: %v", err))
	}
	fmt.Println("Database connection successful!")

	dbGoose, err := dbmigrations.RunDBMigrations(&dbmigrations.Config{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		DB:       cfg.Postgres.DB,
	})
	if err != nil {
		panic(fmt.Errorf("error run DB migrations: %v", err))
	}

	if err := goose.SetDialect("postgres"); err != nil {
		panic(fmt.Errorf("error setting goose dialect: %v", err))
	}

	currentVersion, err := goose.GetDBVersion(dbGoose)
	if err != nil {
		fmt.Printf("Migration table initialization: %v\n", err)
	}

	fmt.Printf("Current migration version: %d\n", currentVersion)

	if err := goose.Up(dbGoose, constants.MigrationFolder); err != nil {
		panic(fmt.Errorf("error running migrations: %v", err))
	}
}
