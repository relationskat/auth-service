package config

import "time"

type GRPC struct {
	Host            string
	Port            string
	RequestTimeout  time.Duration
	ResponseTimeout time.Duration
}
