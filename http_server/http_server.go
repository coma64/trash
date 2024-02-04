package http_server

import (
	"database/sql"
	"errors"
	"github.com/coma64/trash/config"
	"github.com/coma64/trash/db"
	"github.com/coma64/trash/http_server/templates"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"strings"
)

type Server struct {
	echo          *echo.Echo
	db            db.Db
	listenAddress string
}

func NewServer(db db.Db) *Server {
	e := echo.New()
	e.HideBanner = true
	e.Debug = config.Config.Serve.Debug
	e.HTTPErrorHandler = func(err error, c echo.Context) {
		log.Err(err).Msg("Request failed")
		e.DefaultHTTPErrorHandler(err, c)
	}

	server := &Server{
		echo:          e,
		db:            db,
		listenAddress: config.Config.Serve.HttpServerAddress,
	}

	e.Static("/static", "static")
	e.GET("/s/:id", server.GetSnippet)

	return server
}

func (s *Server) Serve() error {
	return s.echo.Start(s.listenAddress)
}

func (s *Server) GetSnippet(c echo.Context) error {
	id := c.Param("id")

	snippet, err := s.db.GetSnippet(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			if err = templates.SnippetNotFound().Render(c.Request().Context(), c.Response().Writer); err != nil {
				log.Err(err).Msg("Failed to render snippet not found")
				return err
			}

			return nil
		}

		return err
	}

	userAgent := c.Request().Header.Get("User-Agent")
	raw := c.QueryParam("raw")
	if raw == "true" || (raw == "" && (strings.HasPrefix(userAgent, "curl") || strings.HasPrefix(userAgent, "Wget"))) {
		return c.Blob(200, "text/plain", []byte(snippet.Content))
	}

	if err = templates.Snippet(snippet, false).Render(c.Request().Context(), c.Response().Writer); err != nil {
		log.Err(err).Msg("Failed to render snippet")
		return err
	}

	return nil
}
