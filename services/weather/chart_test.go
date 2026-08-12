package weather_test

import (
	"strings"
	"testing"
	"time"

	"github.com/jonathanribas/epaperbackend/services/weather"
)

func TestBuildHourlyChartHidesWindAndHumidityByDefault(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 10, 22, 0, 0, 0, time.UTC)
	hourly := make([]weather.HourlyPoint, 12)
	for i := range hourly {
		hourly[i] = weather.HourlyPoint{
			Time:         base.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			TemperatureC: 20 + float64(i),
			WindSpeedKmh: 10,
			HumidityPct:  40 + i,
			WeatherCode:  0,
		}
	}
	current := weather.CurrentResponse{WindSpeedKmh: 5, WeatherCode: 0}
	svg := weather.BuildHourlyChart(hourly, weather.SunResponse{}, base, time.UTC, current)
	if strings.Contains(svg, "km/h") || strings.Contains(svg, "40%") {
		t.Fatal("expected wind/humidity hidden when calm and dry")
	}
	if !strings.Contains(svg, "20°") {
		t.Fatal("expected temperature labels")
	}
}

func TestBuildHourlyChartShowsWindAndHumidityWhenRelevant(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 10, 22, 0, 0, 0, time.UTC)
	hourly := make([]weather.HourlyPoint, 12)
	for i := range hourly {
		hourly[i] = weather.HourlyPoint{
			Time:         base.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			TemperatureC: 20 + float64(i),
			WindSpeedKmh: 35,
			HumidityPct:  70,
			WeatherCode:  61,
		}
	}
	sunrise := base.Add(8 * time.Hour).Format(time.RFC3339)
	current := weather.CurrentResponse{WindSpeedKmh: 35, WeatherCode: 61, HumidityPct: 70}
	svg := weather.BuildHourlyChart(hourly, weather.SunResponse{NextSunrise: sunrise}, base, time.UTC, current)
	for _, want := range []string{"35 km/h", "70%", "↑ lever", "stroke-dasharray"} {
		if !strings.Contains(svg, want) {
			t.Fatalf("expected %q in chart svg", want)
		}
	}
}

func TestIsRainCode(t *testing.T) {
	t.Parallel()
	if !weather.IsRainCode(61) || weather.IsRainCode(0) {
		t.Fatal("rain code mapping failed")
	}
}

func TestBuildHourlyChartSparseLabels(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 8, 10, 22, 0, 0, 0, time.UTC)
	temps := []float64{20, 21, 22, 28, 24, 23, 12, 15, 16, 17, 18, 19} // max@3=28, min@6=12
	hourly := make([]weather.HourlyPoint, len(temps))
	for i, temp := range temps {
		hourly[i] = weather.HourlyPoint{
			Time:         base.Add(time.Duration(i) * time.Hour).Format(time.RFC3339),
			TemperatureC: temp,
			WindSpeedKmh: 5,
			HumidityPct:  40,
			WeatherCode:  0,
		}
	}
	svg := weather.BuildHourlyChartWithOptions(hourly, weather.SunResponse{}, base, time.UTC, weather.CurrentResponse{}, weather.ChartOptions{
		SparseLabels: true,
		TempOnly:     true,
	})
	for _, want := range []string{"20°", "19°", "28°", "12°"} {
		if !strings.Contains(svg, want) {
			t.Fatalf("expected sparse label %q", want)
		}
	}
	// Middle non-extrema should not be labeled (e.g. 24° at index 4).
	if strings.Count(svg, "24°") != 0 {
		t.Fatal("expected 24° unlabeled in sparse mode")
	}
	if strings.Contains(svg, "km/h") {
		t.Fatal("temp-only chart should hide wind")
	}
}
