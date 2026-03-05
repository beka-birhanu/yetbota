package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
	"github.com/beka-birhanu/yetbota/moderation-service/drivers/validator"
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

	// Validate the configuration
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

	// Replace placeholders like ${VAR_NAME} with environment variables after reading the config
	replacePlaceholdersWithEnv(config)

	return config
}

// This function will iterate through all keys in the config and replace any placeholders like ${VAR_NAME} with their environment values
func replacePlaceholdersWithEnv(config *viper.Viper) {
	// Retrieve all keys in the configuration
	for _, key := range config.AllKeys() {
		// Get the value for each key
		value := config.GetString(key)

		// Check if the value contains a placeholder like ${VAR_NAME}
		if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
			// Extract the environment variable name from the placeholder
			envVar := strings.TrimSuffix(strings.TrimPrefix(value, "${"), "}")

			// Get the value of the environment variable
			envValue := os.Getenv(envVar)

			// If the environment variable is not found, panic
			if envValue == "" {
				log.Fatalf("Mandatory environment variable %s not found", envVar)
			}

			// Replace the placeholder in the config with the environment variable's value
			config.Set(key, envValue)
		}
	}
}
