# SKILL — Ajouter un service epaperbackend

Ce guide décrit comment créer un nouveau service dans le monorepo `epaperbackend`.

## 1. Créer le répertoire

```
services/<name>/
├── service.go       # implémente service.Service
├── handler.go       # routes HTTP (optionnel si tout dans service.go)
├── README.md
└── plugin/
    ├── settings.yml
    ├── shared.liquid (optionnel)
    ├── full.liquid
    ├── half_vertical.liquid (selon besoin)
    └── quadrant.liquid (selon besoin)
```

## 2. Implémenter l'interface

```go
type Service interface {
    Name() string
    RoutePrefix() string
    Register(mux *http.ServeMux) error
    Health(ctx context.Context) HealthStatus
    EnvPrefix() string
    PluginDir() string
    NeedsCache() bool
}
```

Enregistrer via factory dans `init()` :

```go
func init() {
    registry.RegisterFactory(func(cfg config.Config) (service.Service, error) {
        return New(cfg)
    }, "myname")
}
```

Puis blank import dans `server/main.go` :

```go
_ "github.com/jonathanribas/epaperbackend/services/myname"
```

## 3. Configuration env

Préfixe : `SERVICE_<ENV_PREFIX>_<KEY>`

Exemples :
- `SERVICE_WEATHER_CACHE_TTL_MINUTES=30`
- `SERVICE_RSS_FEEDS=https://example.com/feed.xml`

Helpers : `config.GetServiceString`, `config.GetServiceInt`.

## 4. Cache SQLite (si `NeedsCache() == true`)

```go
store, err := cache.Open(cfg.DataDir, "myname", 1, cache.DefaultMigrate)
```

- DB : `{DATA_DIR}/myname.db`
- Clé de cache stable (hash des paramètres de requête)
- TTL configurable via env

Services **sans cache** (ex: NTP) : `NeedsCache() false`, pas de DB.

## 5. Plugin Larapaper

Format recipe importable (ZIP) :

- `settings.yml` avec `strategy: polling` et `polling_url` utilisant `{{ config_fields }}`
- Templates Liquid avec variables `data`, `size`, `config`
- Tailles Larapaper : `full`, `half_vertical`, `quadrant`

Tester via UI debug (`make debug`) → http://127.0.0.1:4242/debug/<name>

Champs résolution preview : width/height (défaut 800×480), persistés en localStorage.

## 6. Tests

- Unitaires : logique métier, mapping, cache
- Handler : `httptest` + mocks upstream
- Preview : `server/debug/preview` si templates Liquid
- Lancer : `go test ./...`

## 7. Patrons par type de service

| Type | Exemple | Cache | Notes |
|------|---------|-------|-------|
| API externe | météo | oui | TTL minutes/heures |
| Scraping | MotoGP | oui | refresh 1×/jour, respect robots.txt |
| Stateless | NTP | non | réponse calculée à la volée |
| Agrégation | RSS | oui | merge feeds, limiter N items |
| Média | XKCD, Civitai | oui | URL image + métadonnées |
| AI agent | résumé Wikipedia | oui | appel agent + cache long |

## 8. Checklist PR

- [ ] Service enregistré et routes OK
- [ ] README service avec curl + env vars
- [ ] Plugin Larapaper dans `plugin/`
- [ ] Tests passent
- [ ] UI debug liste le service
