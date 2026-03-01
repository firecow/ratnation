package debugcmd

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
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

func runRequester(ctx context.Context, rawURL string, interval time.Duration) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	tick(ctx, client, parsedURL)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tick(ctx, client, parsedURL)
		}
	}
}

func tick(ctx context.Context, client *http.Client, targetURL *url.URL) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    targetURL,
		Host:   targetURL.Host,
	}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Request error", "error", err)
		return
	}
	_ = resp.Body.Close()

	statusCode := resp.StatusCode
	if statusCode != http.StatusOK {
		slog.Error("Request failed", "status_code", statusCode)
		return
	}

	slog.Info("Request successful", "status_code", statusCode)
}
