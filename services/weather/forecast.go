package weather

import "time"

var frenchWeekdays = [...]string{
	"dimanche", "lundi", "mardi", "mercredi", "jeudi", "vendredi", "samedi",
}

type LocationResponse struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type CurrentResponse struct {
	TemperatureC float64 `json:"temperature_c"`
	WindSpeedKmh float64 `json:"wind_speed_kmh"`
	HumidityPct  int     `json:"humidity_pct"`
	WeatherCode  int     `json:"weather_code"`
	Icon         string  `json:"icon"`
	IconURL      string  `json:"icon_url"`
}

type HourlyPoint struct {
	Time         string  `json:"time"`
	TemperatureC float64 `json:"temperature_c"`
	WindSpeedKmh float64 `json:"wind_speed_kmh"`
	HumidityPct  int     `json:"humidity_pct"`
	WeatherCode  int     `json:"weather_code"`
	Icon         string  `json:"icon"`
}

type SunResponse struct {
	NextSunrise   string `json:"next_sunrise"`
	NextSunset    string `json:"next_sunset"`
	NextSunriseHM string `json:"next_sunrise_hm"`
	NextSunsetHM  string `json:"next_sunset_hm"`
}

type DailyForecast struct {
	Date         string  `json:"date"`
	DayName      string  `json:"day_name"`
	TempMinC     float64 `json:"temp_min_c"`
	TempMaxC     float64 `json:"temp_max_c"`
	WindSpeedKmh float64 `json:"wind_speed_kmh"`
	HumidityPct  int     `json:"humidity_pct"`
	WeatherCode  int     `json:"weather_code"`
	Icon         string  `json:"icon"`
	IconURL      string  `json:"icon_url"`
}

type Response struct {
	Location             LocationResponse `json:"location"`
	FetchedAt            string           `json:"fetched_at"`
	Cached               bool             `json:"cached"`
	Current              CurrentResponse  `json:"current"`
	HourlyNext12         []HourlyPoint    `json:"hourly_next_12"`
	HourlyChartSVG       string           `json:"hourly_chart_svg"`
	HourlyChartSVGSparse string           `json:"hourly_chart_svg_sparse"`
	Sun                  SunResponse      `json:"sun"`
	DailyNext7           []DailyForecast  `json:"daily_next_7"`
}

func BuildResponse(loc *Location, raw *ForecastRaw, now time.Time) *Response {
	locTZ := raw.Location()
	currentIcon := IconName(raw.Current.WeatherCode)

	resp := &Response{
		Location: LocationResponse{
			Name:      loc.Name,
			Country:   loc.Country,
			Latitude:  loc.Latitude,
			Longitude: loc.Longitude,
		},
		FetchedAt: now.Format(time.RFC3339),
		Cached:    false,
		Current: CurrentResponse{
			TemperatureC: raw.Current.Temperature,
			WindSpeedKmh: raw.Current.WindSpeed,
			HumidityPct:  raw.Current.Humidity,
			WeatherCode:  raw.Current.WeatherCode,
			Icon:         currentIcon,
			IconURL:      "/weather/icons/" + currentIcon + ".svg",
		},
	}

	resp.HourlyNext12 = buildHourlyNext12(raw, now, locTZ)
	resp.Sun = buildSun(raw, now, locTZ)
	resp.DailyNext7 = buildDailyNext7(raw, locTZ)
	resp.HourlyChartSVG = BuildHourlyChart(resp.HourlyNext12, resp.Sun, now, locTZ, resp.Current)
	resp.HourlyChartSVGSparse = BuildHourlyChartWithOptions(resp.HourlyNext12, resp.Sun, now, locTZ, resp.Current, ChartOptions{
		SparseLabels: true,
		TempOnly:     true,
	})

	return resp
}

func buildHourlyNext12(raw *ForecastRaw, now time.Time, loc *time.Location) []HourlyPoint {
	var out []HourlyPoint
	start := now.In(loc).Truncate(time.Hour)

	for i, t := range raw.Hourly.Time {
		parsed, err := time.ParseInLocation("2006-01-02T15:04", t, loc)
		if err != nil {
			continue
		}
		if parsed.Before(start) {
			continue
		}
		icon := IconName(safeInt(raw.Hourly.WeatherCode, i))
		out = append(out, HourlyPoint{
			Time:         parsed.Format(time.RFC3339),
			TemperatureC: safeFloat(raw.Hourly.Temperature, i),
			WindSpeedKmh: safeFloat(raw.Hourly.WindSpeed, i),
			HumidityPct:  safeInt(raw.Hourly.Humidity, i),
			WeatherCode:  safeInt(raw.Hourly.WeatherCode, i),
			Icon:         icon,
		})
		if len(out) >= 12 {
			break
		}
	}
	return out
}

func buildSun(raw *ForecastRaw, now time.Time, loc *time.Location) SunResponse {
	var nextSunrise, nextSunset time.Time

	for _, s := range raw.Daily.Sunrise {
		t, err := parseOpenMeteoLocalTime(s, loc)
		if err != nil {
			continue
		}
		if t.After(now) && (nextSunrise.IsZero() || t.Before(nextSunrise)) {
			nextSunrise = t
		}
	}
	for _, s := range raw.Daily.Sunset {
		t, err := parseOpenMeteoLocalTime(s, loc)
		if err != nil {
			continue
		}
		if t.After(now) && (nextSunset.IsZero() || t.Before(nextSunset)) {
			nextSunset = t
		}
	}

	out := SunResponse{}
	if !nextSunrise.IsZero() {
		out.NextSunrise = nextSunrise.Format(time.RFC3339)
		out.NextSunriseHM = nextSunrise.Format("15:04")
	}
	if !nextSunset.IsZero() {
		out.NextSunset = nextSunset.Format(time.RFC3339)
		out.NextSunsetHM = nextSunset.Format("15:04")
	}
	return out
}

func normalizeSunTimes(sun *SunResponse) {
	if sun == nil {
		return
	}
	if sun.NextSunriseHM == "" && sun.NextSunrise != "" {
		if t, err := time.Parse(time.RFC3339, sun.NextSunrise); err == nil {
			sun.NextSunriseHM = t.Format("15:04")
		}
	}
	if sun.NextSunsetHM == "" && sun.NextSunset != "" {
		if t, err := time.Parse(time.RFC3339, sun.NextSunset); err == nil {
			sun.NextSunsetHM = t.Format("15:04")
		}
	}
}

func buildDailyNext7(raw *ForecastRaw, loc *time.Location) []DailyForecast {
	var out []DailyForecast
	for i, date := range raw.Daily.Time {
		if i >= 7 {
			break
		}
		code := safeInt(raw.Daily.WeatherCode, i)
		icon := IconName(code)
		out = append(out, DailyForecast{
			Date:         date,
			DayName:      frenchDayName(date, loc),
			TempMinC:     safeFloat(raw.Daily.TempMin, i),
			TempMaxC:     safeFloat(raw.Daily.TempMax, i),
			WindSpeedKmh: safeFloat(raw.Daily.WindMax, i),
			HumidityPct:  safeInt(raw.Daily.HumidityAvg, i),
			WeatherCode:  code,
			Icon:         icon,
			IconURL:      "/weather/icons/" + icon + ".svg",
		})
	}
	return out
}

func frenchDayName(date string, loc *time.Location) string {
	t, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return date
	}
	return frenchWeekdays[t.Weekday()]
}

func parseOpenMeteoLocalTime(value string, loc *time.Location) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		time.RFC3339,
	}
	var lastErr error
	for _, layout := range layouts {
		t, err := time.ParseInLocation(layout, value, loc)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func safeFloat(values []float64, i int) float64 {
	if i < 0 || i >= len(values) {
		return 0
	}
	return values[i]
}

func safeInt(values []int, i int) int {
	if i < 0 || i >= len(values) {
		return 0
	}
	return values[i]
}
