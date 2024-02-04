package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/coma64/trash/config"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"time"
)

type SqliteDb struct {
	db *sqlx.DB
}

var _ Db = &SqliteDb{}

func NewSqliteDb(dbPath string) (*SqliteDb, error) {
	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	return &SqliteDb{
		db: db,
	}, nil
}

func (s *SqliteDb) InsertSnippet(ctx context.Context, title, content string) (*Snippet, error) {
	id, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("generating uuid: %w", err)
	}

	binaryId, err := id.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshalling uuid: %w", err)
	}

	snippet := Snippet{
		Id:      id.String(),
		Title:   title,
		Content: content,
	}
	if err = s.db.GetContext(
		ctx,
		&snippet,
		"insert into snippets(id, title, content) values ($1, $2, $3) returning created_at",
		binaryId,
		title,
		content,
	); err != nil {
		return nil, fmt.Errorf("inserting snippet: %w", err)
	}

	return &snippet, nil
}

func (s *SqliteDb) GetSnippet(ctx context.Context, id string) (*Snippet, error) {
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("parsing id: %w", err)
	}

	binaryId, err := parsedId.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshalling id: %w", err)
	}

	var snippet Snippet
	if err = s.db.GetContext(ctx, &snippet, "select id, content, created_at from snippets where id = $1", binaryId); err != nil {
		return nil, fmt.Errorf("getting snippet: %w", err)
	}

	if snippet.CreatedAt.Add(config.Config.Serve.SnippetRetentionTime).Before(time.Now()) {
		if err = s.DeleteSnippet(ctx, id); err != nil {
			return nil, fmt.Errorf("deleting expired snippet: %w", err)
		}

		return nil, sql.ErrNoRows
	}

	snippet.Id = parsedId.String()

	return &snippet, nil
}

func (s *SqliteDb) DeleteExpiredSnippets(ctx context.Context) (err error) {
	oldestCreationDate := time.Now().UTC().Add(-config.Config.Serve.SnippetRetentionTime)
	log.Debug().Msgf("Deleting snippets older than %s", oldestCreationDate)

	if _, err = s.db.ExecContext(ctx, "delete from snippets where created_at < $1", oldestCreationDate); err != nil {
		return fmt.Errorf("deleting expired snippets: %w", err)
	}

	return nil
}

func (s *SqliteDb) DeleteSnippet(ctx context.Context, id string) error {
	if _, err := s.db.ExecContext(ctx, "delete from snippets where id = $1", id); err != nil {
		return fmt.Errorf("deleting snippet: %w", err)
	}

	return nil
}

func (s *SqliteDb) ApplyMigrations(embeddedMigrations *embed.FS) error {
	goose.SetBaseFS(embeddedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("setting db dialect: %w", err)
	}

	if err := goose.Up(s.db.DB, "migrations"); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}

	return nil
}
