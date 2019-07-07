package config

import (
	"testing"
)

func TestReadConfigSuccess(t *testing.T) {
	config, err := ReadConfig()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if config.Server.Addr == "" {
		t.Fatalf("failed test server address is nil")
	}

	if config.Server.Port == "" {
		t.Fatalf("failed test server port is nil")
	}
}
