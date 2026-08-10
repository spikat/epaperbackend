package weather

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonathanribas/epaperbackend/pkg/cache"
	"github.com/jonathanribas/epaperbackend/pkg/config"
	"github.com/jonathanribas/epaperbackend/pkg/httpx"
	"github.com/jonathanribas/epaperbackend/pkg/registry"
	"github.com/jonathanribas/epaperbackend/pkg/service"
)

const (
	serviceName    = "weather"
	schemaVersion  = 1
	defaultCountry = "FR"
	defaultCity    = "Marseille"
)

type Service struct {
	cfg    config.Config
	client *Client
	store  *cache.Store
}

func New(cfg config.Config) (*Service, error) {
	store, err := cache.Open(cfg.DataDir, serviceName, schemaVersion, cache.DefaultMigrate)
	if err != nil {
		return nil, err
	}
	return &Service{
		cfg:    cfg,
		client: NewClient(http.DefaultClient),
		store:  store,
	}, nil
}

func init() {
	registry.RegisterFactory(func(cfg config.Config) (service.Service, error) {
		return New(cfg)
	}, serviceName)
}

func (s *Service) Name() string        { return serviceName }
func (s *Service) RoutePrefix() string { return "/weather" }
func (s *Service) EnvPrefix() string   { return "WEATHER" }
func (s *Service) PluginDir() string   { return filepath.Join("services", serviceName, "plugin") }
func (s *Service) NeedsCache() bool    { return true }

func (s *Service) Register(mux *http.ServeMux) error {
	mux.HandleFunc("GET /weather", s.handleWeather)
	mux.HandleFunc("GET /weather/icons/", s.handleIcon)
	return nil
}

func (s *Service) Health(_ context.Context) service.HealthStatus {
	if s.store == nil {
		return service.HealthStatus{OK: false, Message: "cache unavailable"}
	}
	return service.HealthStatus{OK: true, Message: "ready"}
}

func (s *Service) Close() error {
	if s.store == nil {
		return nil
	}
	return s.store.Close()
}

type queryParams struct {
	Country string
	City    string
}

func (s *Service) parseQuery(r *http.Request) queryParams {
	q := queryParams{
		Country: strings.TrimSpace(r.URL.Query().Get("country")),
		City:    strings.TrimSpace(r.URL.Query().Get("city")),
	}
	if q.City == "" {
		q.City = strings.TrimSpace(r.URL.Query().Get("postal"))
	}
	if q.Country == "" {
		q.Country = config.GetServiceString(serviceName, "COUNTRY", defaultCountry)
	}
	if q.City == "" {
		q.City = config.GetServiceString(serviceName, "CITY", defaultCity)
	}
	return q
}

func cacheKey(q queryParams) string {
	raw := strings.ToUpper(q.Country) + "|" + strings.ToLower(q.City)
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *Service) handleWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	q := s.parseQuery(r)
	if q.Country == "" || q.City == "" {
		httpx.WriteError(w, http.StatusBadRequest, "country and city/postal are required")
		return
	}

	ctx := r.Context()
	ttl := time.Duration(config.GetServiceInt(serviceName, "CACHE_TTL_MINUTES", 30)) * time.Minute
	key := cacheKey(q)

	entry, err := s.store.Get(ctx, key)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	now := time.Now()
	if entry != nil && s.store.IsValid(entry, now.UTC()) {
		var cached Response
		if err := json.Unmarshal(entry.Payload, &cached); err == nil {
			normalizeSunTimes(&cached.Sun)
			cached.Cached = true
			payload, err := json.Marshal(cached)
			if err == nil {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache", "HIT")
				_, _ = w.Write(payload)
				return
			}
		}
	}

	resp, err := s.fetchForecast(ctx, q)
	if err != nil {
		httpx.WriteError(w, http.StatusBadGateway, err.Error())
		return
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := s.store.Set(ctx, key, payload, ttl); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	_, _ = w.Write(payload)
}

func (s *Service) fetchForecast(ctx context.Context, q queryParams) (*Response, error) {
	loc, err := s.client.Geocode(ctx, q.City, q.Country)
	if err != nil {
		return nil, err
	}

	raw, err := s.client.Forecast(ctx, loc.Latitude, loc.Longitude, loc.Timezone)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(raw.Location())
	return BuildResponse(loc, raw, now), nil
}

func (s *Service) handleIcon(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/weather/icons/")
	name = strings.TrimSuffix(name, ".svg")
	name = strings.Trim(name, "/")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	svg, ok := IconSVG(name)
	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	_, _ = w.Write([]byte(svg))
}

var _ service.Service = (*Service)(nil)
