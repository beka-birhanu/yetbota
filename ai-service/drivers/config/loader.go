package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/ai-service/drivers/validator"
	"github.com/spf13/viper"
)

func Load() (*Configs, error) {
	configuration := newConfig()

	var appConfig Configs
	err := configuration.Unmarshal(&appConfig)
	if err != nil {
		return nil, &toddlerr.Error{
			PublicStatusCode:  status.ServerError,
			ServiceStatusCode: status.ServerError,
			PublicMessage:     "Failed to load configuration",
			ServiceMessage:    fmt.Sprintf("Error unmarshaling config: %v", err),
		}
	}

	if err := validator.Validate.Struct(&appConfig); err != nil {
		return nil, toddlerr.FromValidationErrors(err)
	}

	return &appConfig, nil
}

func newConfig() *viper.Viper {
	config := viper.New()
	config.SetConfigType("yaml")
	config.SetConfigName("config")
	config.AddConfigPath(".")
	if err := config.ReadInConfig(); err != nil {
		log.Fatalf("got an error reading file config, error: %s", err)
	}

	replacePlaceholdersWithEnv(config)

	return config
}

func replacePlaceholdersWithEnv(config *viper.Viper) {
	for _, key := range config.AllKeys() {
		value := config.GetString(key)

		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			envVar := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")

			envValue := os.Getenv(envVar)

			if envValue == "" {
				log.Fatalf("Mandatory environment variable %s not found", envVar)
			}

			config.Set(key, envValue)
		}
	}
}