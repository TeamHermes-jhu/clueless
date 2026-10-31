// Package server will contain the main server and routing information
package server

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/TeamHermes-jhu/clueless/backend/session"
)

const (
	maxSessionsCount = 20
	bufferSize       = 1024

	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
)

// Server contains the server object and maintains
// a list of sessions.
type Server struct {
	mu       sync.Mutex
	mux      *chi.Mux
	srv      *http.Server
	logger   *slog.Logger
	sessions map[uuid.UUID]*session.Session
}

// New returns a Server object and sets up middleware and routes.
func New(l *slog.Logger) *Server {
	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(slogMiddleware(l))
	mux.Use(middleware.Recoverer)
	mux.Use(corsMiddleware("http://localhost:8080"))

	s := &Server{
		mux:      mux,
		logger:   l,
		sessions: make(map[uuid.UUID]*session.Session),
	}

	s.setupRoutes()

	return s
}

// Serve starts the server on the specified port and address.
func (s *Server) Serve(addr, port string) error {
	s.logger.Info("Starting server on", "address", addr, "port", port)

	s.srv = &http.Server{
		Addr:              addr + ":" + port,
		Handler:           s.mux,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	return s.srv.ListenAndServe()
}

// Shutdown initiates graceful shutdown of the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.InfoContext(ctx, "Got shutdown signal")

	for _, sess := range s.sessions {
		sess.Shutdown(ctx)
	}

	if s.srv != nil {
		return s.srv.Shutdown(ctx)
	}

	return nil
}

func (s *Server) setupRoutes() {
	s.mux.Post("/join_session", s.joinSessionHandler)
	s.mux.Get("/ws", s.handleWS)
}
