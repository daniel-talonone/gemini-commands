package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/daniel-talonone/gemini-commands/internal/server"
	"github.com/spf13/cobra"
)

var servePort int

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 1004, "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the read-only dashboard web server",
	Long: `Starts a persistent HTTP server that displays all ai-session features.

The dashboard scans ~/.features/ on every page request and renders
a feature list with the current pipeline step, running state, and last done task.

Filter by repo:   http://localhost:1004/?repo=org/name
Filter by status: http://localhost:1004/?status=running|idle|done

Flags:
  --port  Port to listen on (default 1004)

Shutdown: press CTRL+C for graceful shutdown.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		srv := server.New(servePort, &server.DashboardScanner{})

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		errCh := make(chan error, 1)
		go func() { errCh <- srv.Start() }()

		select {
		case err := <-errCh:
			return err
		case <-ctx.Done():
			fmt.Println("\nShutting down...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return srv.Shutdown(shutdownCtx)
		}
	},
}
