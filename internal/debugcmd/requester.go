// Package debugcmd provides debug utilities for testing HTTP connectivity.
package debugcmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

const defaultIntervalMilliseconds = 500

// Command returns the cobra command for the debug-requester subcommand.
func Command() *cobra.Command {
	var (
		requestURL string
		interval   int
	)

	cmd := &cobra.Command{
		Use:   "debug-requester",
		Short: "Start calling HTTP requests and print status code",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRequester(
				cmd.Context(),
				requestURL,
				time.Duration(interval)*time.Millisecond,
			)
		},
	}

	cmd.Flags().StringVar(
		&requestURL, "url", "", "URL to request",
	)
	cmd.Flags().IntVar(
		&interval,
		"interval",
		defaultIntervalMilliseconds,
		"Ticker interval in milliseconds",
	)

	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func runRequester(ctx context.Context, rawURL string, interval time.Duration) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	transport := http.DefaultTransport

	tick(ctx, transport, parsedURL)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tick(ctx, transport, parsedURL)
		}
	}
}

func tick(ctx context.Context, transport http.RoundTripper, targetURL *url.URL) {
	reqCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL.String(), nil)
	if err != nil {
		slog.Error("Failed to create request", "error", err)

		return
	}

	response, err := transport.RoundTrip(request)
	if err != nil {
		slog.Error("Request error", "error", err)

		return
	}

	_ = response.Body.Close()

	if response.StatusCode != http.StatusOK {
		slog.Error("Request failed")

		return
	}

	slog.Info("Request successful")
}
