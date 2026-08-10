package cache

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Store struct {
	db   *sql.DB
	name string
}

type Entry struct {
	Key       string
	Payload   []byte
	CreatedAt time.Time
	ExpiresAt time.Time
}

func Open(dataDir, serviceName string, targetVersion int, migrate func(*sql.DB) error) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	path := filepath.Join(dataDir, serviceName+".db")
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	store := &Store{db: db, name: serviceName}
	if err := store.ensureSchema(targetVersion, migrate); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) ensureSchema(targetVersion int, migrate func(*sql.DB) error) error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	var version int
	err := s.db.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_version`).Scan(&version)
	if err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	if version == 0 {
		if err := migrate(s.db); err != nil {
			return err
		}
		if _, err := s.db.Exec(`INSERT INTO schema_version(version) VALUES (?)`, targetVersion); err != nil {
			return fmt.Errorf("insert schema version: %w", err)
		}
		return nil
	}

	if version != targetVersion {
		return fmt.Errorf("unsupported schema version %d for %s (expected %d)", version, s.name, targetVersion)
	}
	return nil
}

func DefaultMigrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cache_entries (
			cache_key TEXT PRIMARY KEY,
			payload BLOB NOT NULL,
			created_at TIMESTAMP NOT NULL,
			expires_at TIMESTAMP NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_cache_expires ON cache_entries(expires_at);
	`)
	if err != nil {
		return fmt.Errorf("create cache_entries: %w", err)
	}
	return nil
}

func (s *Store) Get(ctx context.Context, key string) (*Entry, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT cache_key, payload, created_at, expires_at
		FROM cache_entries
		WHERE cache_key = ?
	`, key)

	var e Entry
	if err := row.Scan(&e.Key, &e.Payload, &e.CreatedAt, &e.ExpiresAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("cache get: %w", err)
	}
	return &e, nil
}

func (s *Store) Set(ctx context.Context, key string, payload []byte, ttl time.Duration) error {
	now := time.Now().UTC()
	expires := now.Add(ttl)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO cache_entries(cache_key, payload, created_at, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(cache_key) DO UPDATE SET
			payload = excluded.payload,
			created_at = excluded.created_at,
			expires_at = excluded.expires_at
	`, key, payload, now, expires)
	if err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

func (s *Store) IsValid(entry *Entry, now time.Time) bool {
	if entry == nil {
		return false
	}
	return now.Before(entry.ExpiresAt)
}

func (s *Store) PurgeExpired(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM cache_entries WHERE expires_at <= ?`, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("purge expired: %w", err)
	}
	return nil
}
