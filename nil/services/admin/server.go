package admin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/NilFoundation/nil/nil/common/logging"
	"github.com/rs/zerolog"
)

var defaultTimeout = 30 * time.Second

type adminServer struct {
	mux    *http.ServeMux
	cfg    *ServerConfig
	logger zerolog.Logger
}

func StartAdminServer(ctx context.Context, cfg *ServerConfig, logger zerolog.Logger) error {
	if !cfg.Enabled {
		return nil
	}

	srv := &adminServer{
		mux:    http.NewServeMux(),
		cfg:    cfg,
		logger: logger,
	}

	// GET http:/./set_log_level?level=info
	srv.mux.HandleFunc("/set_log_level", srv.setLogLevel)
	srv.mux.HandleFunc("/set_libp2p_log_level", srv.setLibp2pLogLevel)
	srv.mux.HandleFunc("/ping", srv.ping)

	if err := srv.serve(ctx); err != nil {
		return fmt.Errorf("error starting admin server: %w", err)
	}

	return nil
}

func (s *adminServer) serve(ctx context.Context) error {
	srv := http.Server{Handler: s.mux, ReadHeaderTimeout: defaultTimeout}

	// Drop dangling unix socket if server previously was not stopped correctly.
	if _, err := os.Stat(s.cfg.UnixSocketPath); err == nil {
		conn, err := net.Dial("unix", s.cfg.UnixSocketPath)
		if err != nil {
			var netErr *net.OpError
			if errors.As(err, &netErr) && errors.Is(netErr, syscall.ECONNREFUSED) {
				s.logger.Info().Msgf("Remove unused socket file: %s", s.cfg.UnixSocketPath)
				os.Remove(s.cfg.UnixSocketPath)
			} else {
				return fmt.Errorf("error connecting to socket: %w", err)
			}
		} else {
			// Close if socket is already open.
			// Error "already in use" will be returned by net.Listen.
			conn.Close()
		}
	}

	listener, err := net.Listen("unix", s.cfg.UnixSocketPath)
	if err != nil {
		return err
	}

	defer func() { //nolint:contextcheck
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.logger.Info().Msg("Stopping admin server...")
		_ = srv.Shutdown(shutdownCtx)
		s.logger.Info().Msg("Stopped admin server")
	}()

	go func() {
		s.logger.Info().Msgf("Serving admin handles at `%s`", s.cfg.UnixSocketPath)
		_ = srv.Serve(listener)
	}()

	<-ctx.Done()
	return nil
}

func (s *adminServer) writeResponse(w http.ResponseWriter, err error, okMsg string) {
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "error: %s", err.Error())
	} else {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, okMsg)
	}
}

func (s *adminServer) setLogLevel(w http.ResponseWriter, r *http.Request) {
	lvl := r.URL.Query().Get("level")
	err := logging.TrySetupGlobalLevel(lvl)
	s.writeResponse(w, err, "set to "+lvl)
}

func (s *adminServer) setLibp2pLogLevel(w http.ResponseWriter, r *http.Request) {
	lvl := r.URL.Query().Get("level")
	err := logging.SetLibp2pLogLevel(lvl)
	s.writeResponse(w, err, "set to "+lvl)
}

func (s *adminServer) ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
