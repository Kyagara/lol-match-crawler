package main

import (
	"context"

	"github.com/Kyagara/equinox"
	"github.com/Kyagara/equinox/clients/lol"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog/log"
)

func fetchSummoners(ctx context.Context, client *equinox.Equinox, db *pgxpool.Pool, entries []lol.LeagueItemV4DTO) {
	log.Info().Msg("Fetching summoners")

	for _, entry := range entries {
		err := db.QueryRow(ctx, "SELECT 1 FROM summoner WHERE id = $1;", entry.SummonerID).Scan()
		if rowExists(err) {
			continue
		}

		summoner, err := client.LOL.SummonerV4.BySummonerID(ctx, lol.KR, entry.SummonerID)
		if err != nil {
			log.Error().Err(err).Str("name", entry.SummonerName).Msg("Error getting summoner")
			continue
		}

		_, err = db.Exec(ctx, "INSERT INTO summoner (id, puuid, entry, summoner) VALUES ($1, $2, $3, $4);", entry.SummonerID, summoner.PUUID, entry, summoner)
		if err != nil {
			log.Error().Err(err).Str("name", entry.SummonerName).Msg("Error inserting summoner")
		}

		log.Info().Str("name", entry.SummonerName).Msg("New summoner")
	}
}
