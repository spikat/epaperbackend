# epaperbackend

Backend Go pour plugins Larapaper (displays e-paper).

## Démarrage rapide

```bash
make all
make run          # API :5678
make debug        # API :5678 + UI debug :4242
make test
```

## Docker

Sur le NAS, à côté de Larapaper ([`docker-compose.yml`](docker-compose.yml)) :

```bash
# depuis ce repo (ou copie le service epaperbackend dans ton compose Larapaper)
docker compose build epaperbackend
docker compose up -d
```

| Service | URL |
|---------|-----|
| API | http://truenas.local:5678 |
| Health | http://truenas.local:5678/health |
| Debug UI | http://truenas.local:4242 |
| Depuis Larapaper (même réseau Docker) | `http://epaperbackend:5678` |

Volume `epaperbackend_data` → `/data` (SQLite cache).

Plugin météo : `backend_url` = `http://epaperbackend:5678` (polling depuis le container Larapaper).

```bash
make docker-build
make docker-run   # run local rapide
```

Volume `/data` pour les bases SQLite de cache.

## Structure

- `server/` — serveur HTTP principal + UI debug
- `services/<name>/` — un service par répertoire (code + plugin Larapaper)
- `pkg/` — librairies partagées (registry, cache, config)

## Ajouter un service

Voir [SKILL.md](SKILL.md).

## Service météo

Voir [services/weather/README.md](services/weather/README.md).
