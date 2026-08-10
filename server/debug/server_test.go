package debug_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonathanribas/epaperbackend/pkg/config"
	"github.com/jonathanribas/epaperbackend/pkg/registry"
	"github.com/jonathanribas/epaperbackend/server/debug"
	"github.com/jonathanribas/epaperbackend/services/weather"
)

func TestDebugServicesEndpoint(t *testing.T) {
	cfg := config.Config{DataDir: t.TempDir(), DebugPreviewWidth: 800, DebugPreviewHeight: 480}
	if err := registry.Bootstrap(cfg); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	srv := debug.New(cfg)
	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("json: %v", err)
	}
	if out["services"] == nil {
		t.Fatal("missing services")
	}
}

func TestPreviewEndpointDimensions(t *testing.T) {
	cfg := config.Config{DataDir: t.TempDir(), DebugPreviewWidth: 800, DebugPreviewHeight: 480}
	if err := registry.Bootstrap(cfg); err != nil {
		t.Fatalf("bootstrap: %v", err)
	}

	body := map[string]any{
		"data": map[string]any{
			"location": map[string]any{"name": "Test"},
			"current": map[string]any{
				"temperature_c": 20,
				"wind_speed_kmh": 5,
				"humidity_pct": 40,
				"icon_url": "/weather/icons/clear.svg",
			},
		},
		"size":   "half_vertical",
		"width":  800,
		"height": 480,
	}
	raw, _ := json.Marshal(body)

	srv := debug.New(cfg)
	req := httptest.NewRequest(http.MethodPost, "/api/preview/weather", bytes.NewReader(raw))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d body: %s", rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("json: %v", err)
	}
	if int(out["width"].(float64)) != 400 || int(out["height"].(float64)) != 480 {
		t.Fatalf("unexpected viewport: %+v", out)
	}
}

func TestWeatherServiceRegistered(t *testing.T) {
	cfg := config.Config{DataDir: t.TempDir()}
	svc, err := weather.New(cfg)
	if err != nil {
		t.Fatalf("new weather: %v", err)
	}
	if svc.Name() != "weather" {
		t.Fatalf("name: %s", svc.Name())
	}
}
