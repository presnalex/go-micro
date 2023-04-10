package service

import "time"

var (
	defaultRegisterTTL      = 30 * time.Second
	defaultRegisterInterval = 2 * time.Second // we don't use registry and this saves memory

	defaultClientRetries        = 1
	defaultClientRequestTimeout = 5 * time.Second
	defaultClientPoolSize       = 100
	defaultClientDialTimeout    = 5 * time.Second
	defaultClientPoolTTL        = 60 * time.Second

	defaultTransportTimeout = 15 * time.Second
)
