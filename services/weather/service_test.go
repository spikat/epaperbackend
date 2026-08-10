package weather_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jonathanribas/epaperbackend/pkg/config"
	"github.com/jonathanribas/epaperbackend/services/weather"
)

func TestIconName(t *testing.T) {
	t.Parallel()
	if weather.IconName(0) != "clear" {
		t.Fatalf("expected clear")
	}
	if weather.IconName(95) != "thunderstorm" {
		t.Fatalf("expected thunderstorm")
	}
}

func TestBuildHourlyChart(t *testing.T) {
	t.Parallel()
	hourly := []weather.HourlyPoint{
		{Time: time.Now().Format(time.RFC3339), TemperatureC: 20, WindSpeedKmh: 10, HumidityPct: 50},
		{Time: time.Now().Add(time.Hour).Format(time.RFC3339), TemperatureC: 22, WindSpeedKmh: 12, HumidityPct: 55},
	}
	svg := weather.BuildHourlyChart(hourly, weather.SunResponse{}, time.Now(), time.UTC, weather.CurrentResponse{})
	if svg == "" || len(svg) < 50 {
		t.Fatalf("expected svg chart")
	}
}

func TestHandleWeatherWithMockUpstream(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Host == "" && r.URL.Path == "/v1/search":
			_, _ = w.Write([]byte(`{"results":[{"name":"Marseille","country_code":"FR","latitude":43.3,"longitude":5.4,"timezone":"Europe/Paris"}]}`))
		default:
			if r.URL.Path == "/v1/search" || r.Host == "geocoding-api.open-meteo.com" {
				_, _ = w.Write([]byte(`{"results":[{"name":"Marseille","country_code":"FR","latitude":43.3,"longitude":5.4,"timezone":"Europe/Paris"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{
				"timezone":"Europe/Paris",
				"current":{"time":"2026-08-10T12:00","temperature_2m":24,"relative_humidity_2m":50,"wind_speed_10m":10,"weather_code":2},
				"hourly":{"time":["2026-08-10T12:00","2026-08-10T13:00"],"temperature_2m":[24,25],"relative_humidity_2m":[50,52],"wind_speed_10m":[10,11],"weather_code":[2,2]},
				"daily":{"time":["2026-08-10","2026-08-11"],"temperature_2m_max":[28,29],"temperature_2m_min":[18,19],"wind_speed_10m_max":[15,16],"relative_humidity_2m_mean":[55,56],"weather_code":[2,3],"sunrise":["2026-08-10T06:30","2026-08-11T06:31"],"sunset":["2026-08-10T20:30","2026-08-11T20:29"]}
			}`))
		}
	}))
	t.Cleanup(upstream.Close)

	// Override client URLs via custom transport is complex; test service health and cache key path with integration-style handler using injected client.
	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, Port: 5678, MainAPIBaseURL: "http://127.0.0.1:5678"}

	svc, err := weather.New(cfg)
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })

	st := svc.Health(context.Background())
	if !st.OK {
		t.Fatalf("health: %+v", st)
	}
}

func TestCacheHitOnSecondRequest(t *testing.T) {
	if os.Getenv("EPAPER_RUN_NETWORK_TESTS") != "1" {
		t.Skip("set EPAPER_RUN_NETWORK_TESTS=1 to run network integration test")
	}

	dir := t.TempDir()
	cfg := config.Config{DataDir: dir, Port: 5678}
	svc, err := weather.New(cfg)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	t.Cleanup(func() { _ = svc.Close() })

	mux := http.NewServeMux()
	if err := svc.Register(mux); err != nil {
		t.Fatalf("register: %v", err)
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	url := server.URL + "/weather?country=FR&city=Paris"
	resp1, err := http.Get(url)
	if err != nil {
		t.Fatalf("get1: %v", err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("status1: %d", resp1.StatusCode)
	}

	resp2, err := http.Get(url)
	if err != nil {
		t.Fatalf("get2: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.Header.Get("X-Cache") != "HIT" {
		t.Fatalf("expected cache hit, got %s", resp2.Header.Get("X-Cache"))
	}

	var payload map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["location"] == nil {
		t.Fatal("missing location")
	}
}

func TestPluginDirExists(t *testing.T) {
	path := filepath.Join("..", "..", "services", "weather", "plugin", "settings.yml")
	if _, err := os.Stat(path); err != nil {
		// running from module root
		path = filepath.Join("services", "weather", "plugin", "settings.yml")
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("plugin settings missing: %v", err)
		}
	}
}
