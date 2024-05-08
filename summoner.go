package main

import (
	"context"

	"github.com/Kyagara/equinox/v2"
	"github.com/Kyagara/equinox/v2/clients/lol"
	"github.com/jackc/pgx/v5/pgxpool"
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
			log.Error().Err(err).Str("summonerID", entry.SummonerID).Msg("Error getting summoner")
			continue
		}

		_, err = db.Exec(ctx, "INSERT INTO summoner (id, puuid, entry, summoner) VALUES ($1, $2, $3, $4);", entry.SummonerID, summoner.PUUID, entry, summoner)
		if err != nil {
			log.Error().Err(err).Str("summonerID", entry.SummonerID).Msg("Error inserting summoner")
		}

		log.Info().Str("summonerID", entry.SummonerID).Msg("New summoner")
	}
}
