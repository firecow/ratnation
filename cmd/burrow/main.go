// Package main is the entry point for the burrow service mesh CLI.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/firecow/burrow/internal/council"
	"github.com/firecow/burrow/internal/debugcmd"
	"github.com/firecow/burrow/internal/king"
	"github.com/firecow/burrow/internal/ling"
	"github.com/spf13/cobra"
)

func main() {
	os.Exit(run())
}

func run() int {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	root := &cobra.Command{
		Use:   "burrow",
		Short: "Distributed service mesh with native QUIC tunneling",
	}

	root.AddCommand(council.Command())
	root.AddCommand(king.Command())
	root.AddCommand(ling.Command())
	root.AddCommand(debugcmd.Command())

	root.SetContext(ctx)

	err := root.ExecuteContext(ctx)
	if err != nil {
		return 1
	}

	return 0
}
