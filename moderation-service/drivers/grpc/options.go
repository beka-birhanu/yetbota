package grpc

import "time"

type Options struct {
	Cert       string        `json:"cert"`
	ServerName string        `json:"server_name"`
	Address    string        `json:"address"`
	Timeout    time.Duration `json:"timeout"`
}
