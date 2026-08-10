package main_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonathanribas/epaperbackend/pkg/config"
	"github.com/jonathanribas/epaperbackend/pkg/httpx"
	"github.com/jonathanribas/epaperbackend/pkg/registry"

	_ "github.com/jonathanribas/epaperbackend/services/weather"
)

func TestMainRoutes(t *testing.T) {
	cfg := config.Config{DataDir: t.TempDir(), Port: 5678}
	if err := registry.Bootstrap(cfg); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	if err := registry.RegisterRoutes(mux); err != nil {
		t.Fatalf("register: %v", err)
	}

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	resp, err := http.Get(server.URL + "/health")
	if err != nil {
		t.Fatalf("health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
}
