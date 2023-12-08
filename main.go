package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Kyagara/equinox/clients/lol"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.DebugLevel)

	ctx := context.Background()

	db, err := newDBConnection(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating database connection")
	}
	defer db.Close()

	client, err := newEquinoxClient()
	if err != nil {
		log.Error().Err(err).Msg("Error creating equinox client")
		return
	}

	ctxWithCancel, cancel := context.WithCancel(ctx)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	defer func() {
		log.Info().Msg("Shutting down")
		signal.Stop(stop)
		cancel()
	}()

	go func() {
		select {
		case <-stop:
			cancel()
		case <-ctx.Done():
		}
		<-stop
		os.Exit(1)
	}()

	log.Info().Msg("Fetching challenger league")
	league, err := client.LOL.LeagueV4.ChallengerByQueue(ctx, lol.KR, lol.RANKED_SOLO_5X5)
	if err != nil {
		log.Error().Err(err).Msg("Error getting challenger league")
		return
	}

	fetchSummoners(ctxWithCancel, client, db, league.Entries)
	fetchMatches(ctxWithCancel, client, db)
}
