package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

const (
	keyAppPort                = "APP_PORT"
	keyClickHouseHost         = "CLICKHOUSE_HOST"
	keyClickHousePort         = "CLICKHOUSE_PORT"
	keyClickHouseUser         = "CLICKHOUSE_USER"
	keyClickHousePassword     = "CLICKHOUSE_PASSWORD"
	keyClickHouseDatabase     = "CLICKHOUSE_DATABASE"
	keyConnectRetries         = "APP_CLICKHOUSE_RETRIES"
	keyConnectRetryDelaySec   = "APP_CLICKHOUSE_RETRY_DELAY_SEC"
	keyConnectPingTimeoutSec  = "APP_CLICKHOUSE_PING_TIMEOUT_SEC"
)

type Config struct {
	Port               string
	ClickHouseHost     string
	ClickHousePort     string
	ClickHouseUser     string
	ClickHousePassword string
	ClickHouseDatabase string
	ConnectRetries     int
	ConnectRetryDelay  time.Duration
	ConnectPingTimeout time.Duration
}

func FromEnv() (Config, error) {
	return Load(os.LookupEnv)
}

func Load(lookup func(string) (string, bool)) (Config, error) {
	keys := []string{
		keyAppPort,
		keyClickHouseHost,
		keyClickHousePort,
		keyClickHouseUser,
		keyClickHouseDatabase,
		keyConnectRetries,
		keyConnectRetryDelaySec,
		keyConnectPingTimeoutSec,
	}
	values := make(map[string]string, len(keys))
	for _, key := range keys {
		value, ok := lookup(key)
		if !ok || value == "" {
			return Config{}, fmt.Errorf("environment variable %s is not set", key)
		}
		values[key] = value
	}
	for _, key := range []string{keyAppPort, keyClickHousePort} {
		if err := validatePort(values[key]); err != nil {
			return Config{}, fmt.Errorf("environment variable %s: %w", key, err)
		}
	}
	retries, err := strconv.Atoi(values[keyConnectRetries])
	if err != nil || retries < 1 {
		return Config{}, fmt.Errorf("environment variable %s must be a positive integer", keyConnectRetries)
	}
	retryDelay, err := positiveDuration(values[keyConnectRetryDelaySec])
	if err != nil {
		return Config{}, fmt.Errorf("environment variable %s: %w", keyConnectRetryDelaySec, err)
	}
	pingTimeout, err := positiveDuration(values[keyConnectPingTimeoutSec])
	if err != nil {
		return Config{}, fmt.Errorf("environment variable %s: %w", keyConnectPingTimeoutSec, err)
	}
	password, _ := lookup(keyClickHousePassword)
	return Config{
		Port:               values[keyAppPort],
		ClickHouseHost:     values[keyClickHouseHost],
		ClickHousePort:     values[keyClickHousePort],
		ClickHouseUser:     values[keyClickHouseUser],
		ClickHousePassword: password,
		ClickHouseDatabase: values[keyClickHouseDatabase],
		ConnectRetries:     retries,
		ConnectRetryDelay:  retryDelay,
		ConnectPingTimeout: pingTimeout,
	}, nil
}

func validatePort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("%q is not a number", value)
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("%q must be a port from 1 to 65535", value)
	}
	return nil
}

func positiveDuration(value string) (time.Duration, error) {
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds < 1 {
		return 0, fmt.Errorf("%q must be a positive integer of seconds", value)
	}
	return time.Duration(seconds) * time.Second, nil
}
