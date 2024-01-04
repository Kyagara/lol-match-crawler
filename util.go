package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Kyagara/equinox"
	"github.com/Kyagara/equinox/api"
	"github.com/Kyagara/equinox/cache"
	"github.com/Kyagara/equinox/ratelimit"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

const schema = `
CREATE TABLE IF NOT EXISTS summoner (
    id TEXT PRIMARY KEY,
	puuid TEXT,
	entry JSONB,
    summoner JSONB
);

CREATE TABLE IF NOT EXISTS match (
    id TEXT PRIMARY KEY,
	match JSONB
);

CREATE TABLE IF NOT EXISTS summoner_match (
	summoner_id TEXT,
	match_id TEXT,
	PRIMARY KEY (summoner_id, match_id),
	FOREIGN KEY (summoner_id) REFERENCES summoner(id),
	FOREIGN KEY (match_id) REFERENCES match(id)
);

CREATE TABLE IF NOT EXISTS timeline (
    id TEXT PRIMARY KEY,
	timeline JSONB,
	FOREIGN KEY (id) REFERENCES match(id)
);`

func newEquinoxClient() *equinox.Equinox {
	config := api.EquinoxConfig{
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		Cache:      &cache.Cache{},
		Key:        os.Getenv("EQUINOX_KEY"),
		Retry:      equinox.DefaultRetry(),
		Logger: api.Logger{
			Level:           zerolog.WarnLevel,
			Pretty:          true,
			EnableTimestamp: true,
		},
		RateLimit: ratelimit.NewInternalRateLimit(0.99, time.Second),
	}
	return equinox.NewClientWithConfig(config)
}

func newDBConnection(ctx context.Context) (*pgxpool.Pool, error) {
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(ctx, schema)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func rowExists(err error) bool {
	return !strings.Contains(err.Error(), "no rows in result set") && !errors.Is(err, pgx.ErrNoRows)
}
