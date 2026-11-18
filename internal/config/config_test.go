package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	lookup := func(key string) (string, bool) {
		values := map[string]string{
			"APP_PORT":                        "8080",
			"CLICKHOUSE_HOST":                 "localhost",
			"CLICKHOUSE_PORT":                 "9000",
			"CLICKHOUSE_USER":                 "default",
			"CLICKHOUSE_PASSWORD":             "secret",
			"CLICKHOUSE_DATABASE":             "energosbyt",
			"APP_CLICKHOUSE_RETRIES":          "10",
			"APP_CLICKHOUSE_RETRY_DELAY_SEC":  "1",
			"APP_CLICKHOUSE_PING_TIMEOUT_SEC": "5",
		}
		value, ok := values[key]
		return value, ok
	}
	cfg, err := Load(lookup)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("port = %q", cfg.Port)
	}
	if cfg.ClickHouseHost != "localhost" {
		t.Errorf("host = %q", cfg.ClickHouseHost)
	}
	if cfg.ClickHousePort != "9000" {
		t.Errorf("clickhouse port = %q", cfg.ClickHousePort)
	}
	if cfg.ClickHouseUser != "default" {
		t.Errorf("user = %q", cfg.ClickHouseUser)
	}
	if cfg.ClickHousePassword != "secret" {
		t.Errorf("password = %q", cfg.ClickHousePassword)
	}
	if cfg.ClickHouseDatabase != "energosbyt" {
		t.Errorf("database = %q", cfg.ClickHouseDatabase)
	}
	if cfg.ConnectRetries != 10 {
		t.Errorf("retries = %d", cfg.ConnectRetries)
	}
	if cfg.ConnectRetryDelay.Seconds() != 1 {
		t.Errorf("retry delay = %v", cfg.ConnectRetryDelay)
	}
	if cfg.ConnectPingTimeout.Seconds() != 5 {
		t.Errorf("ping timeout = %v", cfg.ConnectPingTimeout)
	}
}

func TestLoadMissingVariable(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "APP_PORT" {
			return "", false
		}
		return "1", true
	}
	_, err := Load(lookup)
	if err == nil {
		t.Fatal("expected error for missing APP_PORT")
	}
}

func TestLoadEmptyVariable(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "CLICKHOUSE_DATABASE" {
			return "", true
		}
		return "1", true
	}
	_, err := Load(lookup)
	if err == nil {
		t.Fatal("expected error for empty CLICKHOUSE_DATABASE")
	}
}

func TestLoadBadAppPort(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "APP_PORT" {
			return "not-a-port", true
		}
		return "1", true
	}
	_, err := Load(lookup)
	if err == nil {
		t.Fatal("expected error for bad APP_PORT")
	}
}

func TestLoadPortOutOfRange(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "CLICKHOUSE_PORT" {
			return "70000", true
		}
		return "1", true
	}
	_, err := Load(lookup)
	if err == nil {
		t.Fatal("expected error for out of range CLICKHOUSE_PORT")
	}
}

func TestLoadBadRetries(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "APP_CLICKHOUSE_RETRIES" {
			return "0", true
		}
		return "1", true
	}
	_, err := Load(lookup)
	if err == nil {
		t.Fatal("expected error for zero retries")
	}
}

func TestLoadPasswordIsOptional(t *testing.T) {
	lookup := func(key string) (string, bool) {
		if key == "CLICKHOUSE_PASSWORD" {
			return "", false
		}
		return "1", true
	}
	cfg, err := Load(lookup)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.ClickHousePassword != "" {
		t.Errorf("password = %q, want empty", cfg.ClickHousePassword)
	}
}
