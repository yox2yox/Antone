package config

import (
	"testing"
)

func TestReadConfigSuccess(t *testing.T) {
	bridgeConfig, err := ReadBridgeConfig()
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	if bridgeConfig.Server.Addr == "" {
		t.Fatalf("failed test server address is nil")
	}

	if bridgeConfig.Server.Port == "" {
		t.Fatalf("failed test server port is nil")
	}
}
