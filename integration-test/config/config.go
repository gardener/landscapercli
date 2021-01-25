package config

import "time"

type Config struct {
	Kubeconfig string
	LandscaperNamespace string
	MaxRetries int
	SleepTime time.Duration
}