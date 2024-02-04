package db

import (
	"context"
	"fmt"
	"time"
)

type Snippet struct {
	Id           string
	ClonedFromId string
	Title        string
	Content      string
	CreatedAt    time.Time `db:"created_at"`
}

func (s *Snippet) Url(raw bool) string {
	url := fmt.Sprintf("/s/%s", s.Id)
	if raw {
		url += "?raw=true"
	}

	return url
}

type Db interface {
	InsertSnippet(ctx context.Context, title, content string) (snippet *Snippet, err error)
	GetSnippet(ctx context.Context, id string) (snippet *Snippet, err error)
	DeleteSnippet(ctx context.Context, id string) (err error)
	DeleteExpiredSnippets(ctx context.Context) (err error)
}
