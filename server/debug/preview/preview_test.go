package preview_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jonathanribas/epaperbackend/server/debug/preview"
)

func TestViewportSize(t *testing.T) {
	t.Parallel()
	cases := []struct {
		size             string
		w, h             int
		wantW, wantH     int
	}{
		{"full", 800, 480, 800, 480},
		{"half_vertical", 800, 480, 400, 480},
		{"quadrant", 800, 480, 400, 240},
	}
	for _, c := range cases {
		w, h := preview.ViewportSize(c.size, c.w, c.h)
		if w != c.wantW || h != c.wantH {
			t.Fatalf("%s: got %dx%d want %dx%d", c.size, w, h, c.wantW, c.wantH)
		}
	}
}

func TestRenderPlugin(t *testing.T) {
	dir := filepath.Join("..", "..", "services", "weather", "plugin")
	if _, err := os.Stat(dir); err != nil {
		dir = filepath.Join("services", "weather", "plugin")
	}

	html, err := preview.RenderPlugin(dir, "quadrant", preview.Bindings{
		Data: map[string]any{
			"location": map[string]any{"name": "Marseille", "country": "FR"},
			"current": map[string]any{
				"temperature_c":  22.0,
				"wind_speed_kmh": 10.0,
				"humidity_pct":   50,
				"icon":           "partly_cloudy",
				"icon_url":       "/weather/icons/partly_cloudy.svg",
			},
		},
		Size:   "quadrant",
		Width:  800,
		Height: 480,
		Config: map[string]any{"backend_url": "http://127.0.0.1:5678"},
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if html == "" {
		t.Fatal("empty html")
	}
}
