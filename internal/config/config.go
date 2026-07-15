package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// NetFlow collector
	NetflowAddr    string
	NetflowBufSize int

	// API
	APIAddr       string
	APICorsOrigins string

	// Database
	DuckDBPath       string
	DuckDBMaxOpenConns int

	// Retention
	RetentionDays       int
	CleanupIntervalHours int

	// ASN
	ASNDBPath         string
	ASNCacheTTLHours  int
	ASNCacheMaxEntries int

	// Logging
	LogLevel string
}

func Load() (*Config, error) {
	cfg := &Config{
		NetflowAddr:         envStr("NETFLOW_ADDR", "0.0.0.0:2055"),
		NetflowBufSize:      envInt("NETFLOW_BUFFER_SIZE", 65535),
		APIAddr:             envStr("API_ADDR", "0.0.0.0:8080"),
		APICorsOrigins:      envStr("API_CORS_ORIGINS", "*"),
		DuckDBPath:          envStr("DUCKDB_PATH", "./data/netconstat.duckdb"),
		DuckDBMaxOpenConns:  envInt("DUCKDB_MAX_OPEN_CONNS", 4),
		RetentionDays:       envInt("RETENTION_DAYS", 30),
		CleanupIntervalHours: envInt("CLEANUP_INTERVAL_HOURS", 6),
		ASNDBPath:           envStr("ASN_DB_PATH", "./data/asn/GeoLite2-ASN.mmdb"),
		ASNCacheTTLHours:    envInt("ASN_CACHE_TTL_HOURS", 1),
		ASNCacheMaxEntries:  envInt("ASN_CACHE_MAX_ENTRIES", 500000),
		LogLevel:            envStr("LOG_LEVEL", "info"),
	}

	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	if c.NetflowBufSize < 1500 {
		return fmt.Errorf("NETFLOW_BUFFER_SIZE too small: %d", c.NetflowBufSize)
	}
	if c.RetentionDays < 1 {
		return fmt.Errorf("RETENTION_DAYS must be >= 1, got %d", c.RetentionDays)
	}
	return nil
}

func (c *Config) CleanupInterval() time.Duration {
	return time.Duration(c.CleanupIntervalHours) * time.Hour
}

func (c *Config) ASNCacheTTL() time.Duration {
	return time.Duration(c.ASNCacheTTLHours) * time.Hour
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
