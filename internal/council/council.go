package council

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/firecow/ratnation/internal/state"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "council",
		Short: "Start council",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), port)
		},
	}

	cmd.Flags().IntVar(&port, "port", 8080, "Webserver listening port")

	return cmd
}

func run(ctx context.Context, port int) error {
	s := &state.State{
		Revision: 0,
		Services: []state.StateService{},
		Kings:    []state.StateKing{},
		Lings:    []state.StateLing{},
	}

	var mu sync.RWMutex
	hub := newWSHub()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /state", handleGetState(s, &mu))
	mux.HandleFunc("PUT /king", handlePutKing(s, &mu, hub))
	mux.HandleFunc("PUT /ling", handlePutLing(s, &mu, hub))
	mux.HandleFunc("/ws", hub.handleWebSocket)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	cleanerStop := make(chan struct{})
	go startCleaner(s, &mu, hub, cleanerStop)

	go func() {
		<-ctx.Done()
		slog.Info("Shutdown sequence initiated")
		close(cleanerStop)
		hub.closeAll()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	slog.Info("Ready", "port", port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
