package config

import (
	"yox2yox/antone/internal/log2"
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	// パッケージ内のテストの実行
	code := m.Run()
	// 終了処理
	log2.Close()
	// テストの終了コードで exit
	os.Exit(code)
}

func TestReadConfigSuccess(t *testing.T) {
	bridgeConfig, err := ReadWorkerConfig()
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
