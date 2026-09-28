package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadRejectsUnknownKeys ensures YAML configuration remains strict.
func TestLoadRejectsUnknownKeys(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte("database:\n  dsn: postgres://example\noidc:\n  issuer: https://issuer.example\n  audience: kubecoder-api\nunknown: true\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path, nil); err == nil {
		t.Fatal("expected unknown configuration key to fail")
	}
}

// TestLoadDefaults verifies environment loading and default values.
func TestLoadDefaults(t *testing.T) {
	t.Setenv("KUBECODER_DATABASE_DSN", "postgres://example")
	t.Setenv("KUBECODER_OIDC_ISSUER", "https://issuer.example")
	t.Setenv("KUBECODER_OIDC_AUDIENCE", "kubecoder-api")
	t.Setenv("KUBECODER_HTTP_ALLOWED_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173")
	cfg, err := Load("", nil)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTP.Address != ":8080" || cfg.Log.Format != "json" {
		t.Fatalf("unexpected defaults: %#v", cfg)
	}
	if len(cfg.HTTP.AllowedOrigins) != 2 {
		t.Fatalf("allowed origins = %#v", cfg.HTTP.AllowedOrigins)
	}
}
