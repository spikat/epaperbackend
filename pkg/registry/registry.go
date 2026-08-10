package registry

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"

	"github.com/jonathanribas/epaperbackend/pkg/config"
	"github.com/jonathanribas/epaperbackend/pkg/service"
)

type Factory func(cfg config.Config) (service.Service, error)

var (
	mu         sync.RWMutex
	services   []service.Service
	factories  []namedFactory
	bootstrapped bool
)

type namedFactory struct {
	name string
	fn   Factory
}

func Register(s service.Service) {
	mu.Lock()
	defer mu.Unlock()
	for _, existing := range services {
		if existing.Name() == s.Name() {
			panic(fmt.Sprintf("service already registered: %s", s.Name()))
		}
	}
	services = append(services, s)
}

func RegisterFactory(fn Factory, name string) {
	mu.Lock()
	defer mu.Unlock()
	for _, existing := range factories {
		if existing.name == name {
			panic(fmt.Sprintf("service factory already registered: %s", name))
		}
	}
	factories = append(factories, namedFactory{name: name, fn: fn})
}

func Bootstrap(cfg config.Config) error {
	mu.Lock()
	defer mu.Unlock()
	if bootstrapped {
		return nil
	}
	for _, f := range factories {
		s, err := f.fn(cfg)
		if err != nil {
			return fmt.Errorf("bootstrap %s: %w", f.name, err)
		}
		for _, existing := range services {
			if existing.Name() == s.Name() {
				return fmt.Errorf("service already registered: %s", s.Name())
			}
		}
		services = append(services, s)
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name() < services[j].Name()
	})
	bootstrapped = true
	return nil
}

func All() []service.Service {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]service.Service, len(services))
	copy(out, services)
	return out
}

func Get(name string) (service.Service, bool) {
	mu.RLock()
	defer mu.RUnlock()
	for _, s := range services {
		if s.Name() == name {
			return s, true
		}
	}
	return nil, false
}

func RegisterRoutes(mux *http.ServeMux) error {
	for _, s := range All() {
		if err := s.Register(mux); err != nil {
			return fmt.Errorf("register %s: %w", s.Name(), err)
		}
	}
	return nil
}

func Health(ctx context.Context) map[string]service.HealthStatus {
	out := make(map[string]service.HealthStatus)
	for _, s := range All() {
		out[s.Name()] = s.Health(ctx)
	}
	return out
}

func Infos() []service.Info {
	all := All()
	out := make([]service.Info, len(all))
	for i, s := range all {
		out[i] = service.InfoFrom(s)
	}
	return out
}
