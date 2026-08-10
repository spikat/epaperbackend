package service

import (
	"context"
	"net/http"
)

type HealthStatus struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

type Service interface {
	Name() string
	RoutePrefix() string
	Register(mux *http.ServeMux) error
	Health(ctx context.Context) HealthStatus
	EnvPrefix() string
	PluginDir() string
	NeedsCache() bool
}

type Info struct {
	Name        string `json:"name"`
	RoutePrefix string `json:"route_prefix"`
	EnvPrefix   string `json:"env_prefix"`
	PluginDir   string `json:"plugin_dir"`
	NeedsCache  bool   `json:"needs_cache"`
}

func InfoFrom(s Service) Info {
	return Info{
		Name:        s.Name(),
		RoutePrefix: s.RoutePrefix(),
		EnvPrefix:   s.EnvPrefix(),
		PluginDir:   s.PluginDir(),
		NeedsCache:  s.NeedsCache(),
	}
}
