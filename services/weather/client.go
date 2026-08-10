package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	geocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL  = "https://api.open-meteo.com/v1/forecast"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{httpClient: httpClient}
}

type Location struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

type geocodeResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Country   string  `json:"country_code"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timezone  string  `json:"timezone"`
	} `json:"results"`
}

func (c *Client) Geocode(ctx context.Context, query, country string) (*Location, error) {
	u, err := url.Parse(geocodingURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("name", query)
	q.Set("count", "10")
	q.Set("language", "fr")
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	body, err := c.get(ctx, u.String())
	if err != nil {
		return nil, err
	}

	var resp geocodeResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode geocode: %w", err)
	}
	if len(resp.Results) == 0 {
		return nil, fmt.Errorf("location not found: %s (%s)", query, country)
	}

	country = strings.ToUpper(strings.TrimSpace(country))
	for _, r := range resp.Results {
		if strings.EqualFold(r.Country, country) {
			return &Location{
				Name:      r.Name,
				Country:   r.Country,
				Latitude:  r.Latitude,
				Longitude: r.Longitude,
				Timezone:  r.Timezone,
			}, nil
		}
	}

	r := resp.Results[0]
	return &Location{
		Name:      r.Name,
		Country:   r.Country,
		Latitude:  r.Latitude,
		Longitude: r.Longitude,
		Timezone:  r.Timezone,
	}, nil
}

type ForecastRaw struct {
	Timezone string `json:"timezone"`
	Hourly   struct {
		Time        []string  `json:"time"`
		Temperature []float64 `json:"temperature_2m"`
		Humidity    []int     `json:"relative_humidity_2m"`
		WindSpeed   []float64 `json:"wind_speed_10m"`
		WeatherCode []int     `json:"weather_code"`
	} `json:"hourly"`
	Daily struct {
		Time        []string  `json:"time"`
		TempMax     []float64 `json:"temperature_2m_max"`
		TempMin     []float64 `json:"temperature_2m_min"`
		WindMax     []float64 `json:"wind_speed_10m_max"`
		HumidityAvg []int     `json:"relative_humidity_2m_mean"`
		WeatherCode []int     `json:"weather_code"`
		Sunrise     []string  `json:"sunrise"`
		Sunset      []string  `json:"sunset"`
	} `json:"daily"`
	Current struct {
		Time        string  `json:"time"`
		Temperature float64 `json:"temperature_2m"`
		Humidity    int     `json:"relative_humidity_2m"`
		WindSpeed   float64 `json:"wind_speed_10m"`
		WeatherCode int     `json:"weather_code"`
	} `json:"current"`
}

func (f *ForecastRaw) TimezoneLocation() (*time.Location, error) {
	if f.Timezone == "" {
		return time.UTC, nil
	}
	return time.LoadLocation(f.Timezone)
}

func (c *Client) Forecast(ctx context.Context, lat, lon float64, timezone string) (*ForecastRaw, error) {
	u, err := url.Parse(forecastURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%f", lat))
	q.Set("longitude", fmt.Sprintf("%f", lon))
	q.Set("timezone", timezone)
	q.Set("current", "temperature_2m,relative_humidity_2m,wind_speed_10m,weather_code")
	q.Set("hourly", "temperature_2m,relative_humidity_2m,wind_speed_10m,weather_code")
	q.Set("daily", "temperature_2m_max,temperature_2m_min,wind_speed_10m_max,relative_humidity_2m_mean,weather_code,sunrise,sunset")
	q.Set("forecast_days", "8")
	q.Set("wind_speed_unit", "kmh")
	u.RawQuery = q.Encode()

	body, err := c.get(ctx, u.String())
	if err != nil {
		return nil, err
	}

	var raw ForecastRaw
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode forecast: %w", err)
	}
	return &raw, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "epaperbackend/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("upstream status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (f *ForecastRaw) Location() *time.Location {
	loc, err := f.TimezoneLocation()
	if err != nil {
		return time.UTC
	}
	return loc
}
