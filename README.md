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

```bash
make docker-build
make docker-run
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
