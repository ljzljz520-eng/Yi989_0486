package config

import "testing"

func TestConfigDefaults(t *testing.T) {
	cfg := Default()
	if cfg.Port == 0 || cfg.DataPath == "" {
		t.Fatal(cfg)
	}
	if cfg.ListenAddress() == "" {
		t.Fatal("address")
	}
}
