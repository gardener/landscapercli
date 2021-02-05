package util

import "time"

type Config struct {
	Kubeconfig          string
	LandscaperNamespace string
	TestNamespace       string
	MaxRetries          int
	SleepTime           time.Duration
	RegistryBaseURL     string
}
