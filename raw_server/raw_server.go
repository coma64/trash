package raw_server

import (
	"context"
	"fmt"
	"github.com/coma64/trash/config"
	"github.com/coma64/trash/db"
	"github.com/rs/zerolog/log"
	"net"
	"time"
)

type SnippetUrlFn func(id string) string

type Server struct {
	db                        db.Db
	uploadTimeout             time.Duration
	uploadReadBufferSizeBytes int
	uploadContentMaxSizeBytes int
	snippetUrl                SnippetUrlFn
}

func NewServer(db db.Db, snippetUrlFn SnippetUrlFn) *Server {
	uploadTimeout := 30 * time.Second
	if config.Config.Serve.Debug {
		uploadTimeout = time.Hour
	}

	return &Server{
		db:                        db,
		uploadTimeout:             uploadTimeout,
		uploadReadBufferSizeBytes: 2048,
		uploadContentMaxSizeBytes: int(config.Config.Serve.SnippetMaxSizeBytes),
		snippetUrl:                snippetUrlFn,
	}
}

type connection struct {
	connection net.Conn
	server     *Server
}

func newConnection(conn net.Conn, server *Server) *connection {
	return &connection{
		connection: conn,
		server:     server,
	}
}

func (s *Server) Serve() error {
	log.Debug().Str("address", config.Config.Serve.RawServerAddress).Msg("Binding port")

	listener, err := net.Listen("tcp", config.Config.Serve.RawServerAddress)
	if err != nil {
		return fmt.Errorf("binding port: %w", err)
	}

	defer func() {
		if err := listener.Close(); err != nil {
			log.Err(err).Msg("Failed to close socket")
		}
	}()

	for {
		rawConnection, err := listener.Accept()
		if err != nil {
			log.Err(err).Msg("Failed to accept connection")
			continue
		}

		go func() {
			ctx := log.With().Str("remoteAddress", rawConnection.RemoteAddr().String()).Logger().WithContext(context.Background())

			defer func() {
				if err = rawConnection.Close(); err != nil {
					log.Ctx(ctx).Err(err).Msg("Failed to close connection")
				}
			}()

			deadline := time.Now().Add(s.uploadTimeout)
			ctx, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()

			if err = rawConnection.SetDeadline(deadline); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to set deadline on connection")
			}

			if err = newConnection(rawConnection, s).handleConnection(ctx); err != nil {
				log.Ctx(ctx).Err(err).Msg("Failed to handle connection")
			}
		}()
	}
}

func (c *connection) handleConnection(ctx context.Context) error {
	contents, err := c.readContent(ctx)
	if err != nil {
		return fmt.Errorf("reading contents: %w", err)
	}

	snippet, err := c.server.db.InsertSnippet(ctx, "", string(contents))
	if err != nil {
		return fmt.Errorf("inserting snippet: %w", err)
	}

	if _, err = c.connection.Write([]byte(c.server.snippetUrl(snippet.Id))); err != nil {
		return fmt.Errorf("writing snippet url: %w", err)
	}

	return nil
}

var ErrContentTooBig = fmt.Errorf("content too big")

func (c *connection) readContent(ctx context.Context) ([]byte, error) {
	totalBuffer := make([]byte, 0)
	buffer := make([]byte, c.server.uploadReadBufferSizeBytes)

	for {
		select {
		case <-ctx.Done():
			c.tryWriteString(ctx, "ERROR: Timeout. Closing socket")
			return nil, context.DeadlineExceeded
		default:
			bytesRead, err := c.connection.Read(buffer)
			if err != nil {
				if err.Error() == "EOF" {
					return totalBuffer, nil
				}

				c.tryWriteString(ctx, "ERROR: Failed to read. Closing socket")
				return nil, fmt.Errorf("reading contents: %w", err)
			}

			if len(totalBuffer)+len(buffer) > c.server.uploadContentMaxSizeBytes {
				c.tryWriteString(ctx, "ERROR: Content too big. Closing socket")
				return nil, ErrContentTooBig
			}

			totalBuffer = append(totalBuffer, buffer[:bytesRead]...)
		}
	}
}

func (c *connection) tryWriteString(ctx context.Context, message string) {
	log.Ctx(ctx).Debug().Str("message", message).Msg("Trying to write to connection")
	_, _ = c.connection.Write([]byte(message))
}
