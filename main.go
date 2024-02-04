package main

import (
	"context"
	"embed"
	"fmt"
	"github.com/alecthomas/kong"
	"github.com/coma64/trash/config"
	db2 "github.com/coma64/trash/db"
	"github.com/coma64/trash/http_server"
	"github.com/coma64/trash/raw_server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
	"time"
)

//go:embed migrations
var embeddedMigrations embed.FS

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	ctx := kong.Parse(&config.Config)
	switch ctx.Command() {
	case "serve":
		if err := serve(); err != nil {
			log.Fatal().Err(err).Msg("Failed to serve. Exiting")
		}
	default:
		log.Fatal().Str("command", ctx.Command()).Msg("Unknown command")
	}
}

func serve() error {
	db, err := db2.NewSqliteDb(config.Config.Serve.Globals.DbPath)
	if err != nil {
		return fmt.Errorf("creating db: %w", err)
	}

	if config.Config.Serve.AutomaticallyApplyMigrations {
		if err := db.ApplyMigrations(&embeddedMigrations); err != nil {
			return fmt.Errorf("applying migrations: %w", err)
		}
	}

	wg := &sync.WaitGroup{}

	wg.Add(1)
	rawServer := raw_server.NewServer(
		db,
		func(id string) string {
			return fmt.Sprintf(
				"http://%s/s/%s",
				config.Config.Serve.HttpServerAddress,
				id,
			)
		},
	)
	go func() {
		defer wg.Done()

		// TODO: stop everything
		if err := rawServer.Serve(); err != nil {
			log.Err(err).Msg("Failed to serve raw server")
		}
	}()

	wg.Add(1)
	httpServer := http_server.NewServer(db)
	go func() {
		defer wg.Done()

		if err := httpServer.Serve(); err != nil {
			log.Err(err).Msg("Failed to serve http server")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		time.Sleep(10 * time.Second)
		if err := garbageCollector(context.Background(), db); err != nil {
			log.Err(err).Msg("Garbage collector failed")
		}

		ticker := config.Config.Serve.SnippetGarbageCollectionInterval
		for {
			select {
			case <-time.After(ticker):
				if err := garbageCollector(context.Background(), db); err != nil {
					log.Err(err).Msg("Garbage collector failed")
				}
			}
		}
	}()

	wg.Wait()

	return nil
}

func garbageCollector(ctx context.Context, db db2.Db) error {
	log.Debug().Msg("Running snippet garbage collector")

	if err := db.DeleteExpiredSnippets(ctx); err != nil {
		return fmt.Errorf("deleting expired snippets: %w", err)
	}

	return nil
}
