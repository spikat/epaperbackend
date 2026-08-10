package config_test

import (
	"os"
	"testing"

	"github.com/jonathanribas/epaperbackend/pkg/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DEBUG", "")
	t.Setenv("DATA_DIR", "")

	cfg := config.Load()
	if cfg.Port != 5678 {
		t.Fatalf("port: got %d", cfg.Port)
	}
	if cfg.DebugPort != 4242 {
		t.Fatalf("debug port: got %d", cfg.DebugPort)
	}
	if cfg.DebugPreviewWidth != 800 || cfg.DebugPreviewHeight != 480 {
		t.Fatalf("preview defaults: %dx%d", cfg.DebugPreviewWidth, cfg.DebugPreviewHeight)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("DEBUG", "true")
	t.Setenv("DEBUG_PREVIEW_WIDTH", "1024")
	t.Setenv("DEBUG_PREVIEW_HEIGHT", "600")

	cfg := config.Load()
	if cfg.Port != 9000 || !cfg.Debug {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.DebugPreviewWidth != 1024 || cfg.DebugPreviewHeight != 600 {
		t.Fatalf("preview: %dx%d", cfg.DebugPreviewWidth, cfg.DebugPreviewHeight)
	}
}

func TestServiceEnvHelpers(t *testing.T) {
	key := "SERVICE_WEATHER_CACHE_TTL_MINUTES"
	if err := os.Setenv(key, "45"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Unsetenv(key) })

	if got := config.GetServiceInt("weather", "CACHE_TTL_MINUTES", 30); got != 45 {
		t.Fatalf("got %d", got)
	}
	if got := config.GetServiceString("weather", "COUNTRY", "FR"); got != "FR" {
		t.Fatalf("got %s", got)
	}
}
