package main

import (
	"context"

	"github.com/Kyagara/equinox/v2"
	"github.com/Kyagara/equinox/v2/api"
	"github.com/Kyagara/equinox/v2/clients/lol"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func fetchMatches(ctx context.Context, client *equinox.Equinox, db *pgxpool.Pool) {
	log.Info().Msg("Fetching matches and timelines")

	rows, err := db.Query(ctx, "SELECT id, puuid, summoner FROM summoner;")
	if err != nil {
		log.Error().Err(err).Msg("Error querying summoners")
		return
	}

	for rows.Next() {
		var summonerId string
		var puuid string
		var summoner lol.SummonerV4DTO
		err := rows.Scan(&summonerId, &puuid, &summoner)
		if err != nil {
			log.Error().Err(err).Msg("Error scanning summoner")
			continue
		}

		list, err := client.LOL.MatchV5.ListByPUUID(ctx, api.ASIA, summoner.PUUID, -1, -1, 420, "ranked", -1, 5)
		if err != nil {
			log.Error().Str("puuid", summoner.PUUID).Err(err).Msg("Error getting match list")
			continue
		}

		for _, matchID := range list {
			err := db.QueryRow(ctx, "SELECT 1 FROM match WHERE id = $1;", matchID).Scan()
			if rowExists(err) {
				checkSummonerInMatch(ctx, db, summonerId, matchID)
				checkOrInsertTimeline(ctx, client, db, matchID)
				continue
			}

			match, err := client.LOL.MatchV5.ByID(ctx, api.ASIA, matchID)
			if err != nil {
				log.Error().Str("match_id", matchID).Err(err).Msg("Error getting match")
				continue
			}

			_, err = db.Exec(ctx, "INSERT INTO match (id, match) VALUES ($1, $2);", matchID, match)
			if err != nil {
				log.Error().Str("match_id", matchID).Err(err).Msg("Error inserting match")
				continue
			}

			log.Info().Str("match_id", matchID).Msg("New match")

			checkSummonerInMatch(ctx, db, summonerId, matchID)
			checkOrInsertTimeline(ctx, client, db, matchID)
		}
	}
}

func checkOrInsertTimeline(ctx context.Context, client *equinox.Equinox, db *pgxpool.Pool, matchID string) {
	err := db.QueryRow(ctx, "SELECT 1 FROM timeline WHERE id = $1;", matchID).Scan()
	if rowExists(err) {
		return
	}

	timeline, err := client.LOL.MatchV5.Timeline(ctx, api.ASIA, matchID)
	if err != nil {
		log.Error().Str("match_id", matchID).Err(err).Msg("Error getting timeline")
		return
	}

	_, err = db.Exec(ctx, "INSERT INTO timeline (id, timeline) VALUES ($1, $2);", matchID, timeline)
	if err != nil {
		log.Error().Str("match_id", matchID).Err(err).Msg("Error inserting timeline")
		return
	}
}

func checkSummonerInMatch(ctx context.Context, db *pgxpool.Pool, summonerID string, matchID string) {
	err := db.QueryRow(ctx, "SELECT 1 FROM summoner_match WHERE match_id = $1 AND summoner_id = $2;", matchID, summonerID).Scan()
	if rowExists(err) {
		return
	}
	_, err = db.Exec(ctx, "INSERT INTO summoner_match (summoner_id, match_id) VALUES ($1, $2);", summonerID, matchID)
	if err != nil {
		log.Error().Str("summoner_id", summonerID).Str("match_id", matchID).Err(err).Msg("Error inserting summoner_match")
		return
	}
}
