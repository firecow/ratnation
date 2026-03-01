package debug

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	var (
		url      string
		interval int
	)

	cmd := &cobra.Command{
		Use:   "debug-requester",
		Short: "Start calling HTTP requests and print status code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRequester(cmd.Context(), url, time.Duration(interval)*time.Millisecond)
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "URL to request")
	cmd.Flags().IntVar(&interval, "interval", 500, "Ticker interval in milliseconds")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func runRequester(ctx context.Context, url string, interval time.Duration) error {
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	tick(ctx, client, url)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tick(ctx, client, url)
		}
	}
}

func tick(ctx context.Context, client *http.Client, url string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		slog.Error("Request error", "error", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Request error", "error", err)
		return
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("Request failed", "status_code", resp.StatusCode)
		return
	}

	slog.Info("Request successful", "status_code", resp.StatusCode)
}
