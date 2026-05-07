package temporal

import (
	"time"

	"go.temporal.io/sdk/client"
)

type Config struct {
	Host      string
	Namespace string
}

func NewClient(cfg *Config) (client.Client, error) {
	return client.Dial(client.Options{
		HostPort:  cfg.Host,
		Namespace: cfg.Namespace,
		ConnectionOptions: client.ConnectionOptions{
			GetSystemInfoTimeout: 30 * time.Second,
		},
	})
}
