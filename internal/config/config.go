// SPDX-License-Identifier: AGPL-3.0-or-later

package config

import (
	"os"
	"strconv"
	"time"
)

// RateLimitRouteConfig holds configuration for a specific route type
type RateLimitRouteConfig struct {
	Requests int
	Period   time.Duration
	Burst    int
}

// RateLimitConfig holds all rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool
	CleanupInterval time.Duration

	Register RateLimitRouteConfig
	Snapshot RateLimitRouteConfig
	Admin    RateLimitRouteConfig

	BruteForceThreshold int
	BruteForceBan       time.Duration
}

// LoadRateLimitConfig loads rate limiting configuration from environment variables
func LoadRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Enabled:         getEnvBool("SHM_RATELIMIT_ENABLED", true),
		CleanupInterval: getEnvDuration("SHM_RATELIMIT_CLEANUP_INTERVAL", 10*time.Minute),

		Register: RateLimitRouteConfig{
			Requests: getEnvInt("SHM_RATELIMIT_REGISTER_REQUESTS", 5),
			Period:   getEnvDuration("SHM_RATELIMIT_REGISTER_PERIOD", time.Minute),
			Burst:    getEnvInt("SHM_RATELIMIT_REGISTER_BURST", 2),
		},
		Snapshot: RateLimitRouteConfig{
			Requests: getEnvInt("SHM_RATELIMIT_SNAPSHOT_REQUESTS", 1),
			Period:   getEnvDuration("SHM_RATELIMIT_SNAPSHOT_PERIOD", time.Minute),
			Burst:    getEnvInt("SHM_RATELIMIT_SNAPSHOT_BURST", 2),
		},
		Admin: RateLimitRouteConfig{
			Requests: getEnvInt("SHM_RATELIMIT_ADMIN_REQUESTS", 60),
			Period:   getEnvDuration("SHM_RATELIMIT_ADMIN_PERIOD", time.Minute),
			Burst:    getEnvInt("SHM_RATELIMIT_ADMIN_BURST", 20),
		},

		BruteForceThreshold: getEnvInt("SHM_RATELIMIT_BRUTEFORCE_THRESHOLD", 5),
		BruteForceBan:       getEnvDuration("SHM_RATELIMIT_BRUTEFORCE_BAN", 15*time.Minute),
	}
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if val == "true" || val == "1" || val == "yes" {
			return true
		}
		if val == "false" || val == "0" || val == "no" {
			return false
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
