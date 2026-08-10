# Weather service

Backend météo pour Larapaper, alimenté par [Open-Meteo](https://open-meteo.com/) (sans clé API).

## Endpoint

```
GET /weather?country=FR&city=Marseille
GET /weather?country=FR&postal=13001
```

## Variables d'environnement

| Variable | Défaut | Description |
|----------|--------|-------------|
| `SERVICE_WEATHER_CACHE_TTL_MINUTES` | `30` | Durée de cache SQLite |
| `SERVICE_WEATHER_COUNTRY` | `FR` | Pays par défaut |
| `SERVICE_WEATHER_CITY` | `Marseille` | Ville par défaut |

## Exemple

```bash
curl 'http://127.0.0.1:5678/weather?country=FR&city=Marseille'
```

## Plugin Larapaper

Importer le dossier `plugin/` comme ZIP (structure `settings.yml` + templates Liquid).

Champs de configuration plugin :
- `backend_url` — URL du backend (ex: `http://192.168.1.10:5678` sur NAS)
- `country`, `city` — localisation

## Debug UI

Avec `make debug`, ouvrir http://127.0.0.1:4242/debug/weather pour tester l'API, éditer le plugin et générer une preview HTML.
