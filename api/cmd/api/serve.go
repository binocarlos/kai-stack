package goapi

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/binocarlos/kai-stack/api/pkg/jobqueue"
	"github.com/binocarlos/kai-stack/api/pkg/server"
	"github.com/binocarlos/kai-stack/api/pkg/store"
	"github.com/binocarlos/kai-stack/api/pkg/system"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newServeCmd() *cobra.Command {
	serveConfig, err := newConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create serve options")
	}

	envHelpText := generateEnvHelpText(serveConfig, "")

	serveCmd := &cobra.Command{
		Use:     "serve",
		Short:   "Start the platinum api server.",
		Long:    "Start the platinum api server.",
		Example: "TBD",
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := serve(cmd, serveConfig)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to run server")
			}
			return nil
		},
	}

	serveCmd.Long += "\n\nEnvironment Variables:\n\n" + envHelpText

	return serveCmd
}

func serve(cmd *cobra.Command, cfg *config.Config) error {
	system.SetupLogging()

	if !cfg.WebServer.Enabled && !cfg.Worker.Enabled {
		return fmt.Errorf("nothing to run: enable SERVER_ENABLED and/or WORKER_ENABLED")
	}

	// Cleanup manager ensures that resources are freed before exiting:
	cm := system.NewCleanupManager()
	defer cm.Cleanup(cmd.Context())

	// Create a cancellable context for license checks
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	// Context ensures main goroutine waits until killed with ctrl+c:
	ctx, signalCancel := signal.NotifyContext(ctx, os.Interrupt)
	defer signalCancel()

	postgresStore, err := store.NewPostgresStore(cfg.Database)
	if err != nil {
		return err
	}
	cm.RegisterCallback(postgresStore.Close) // close the pool on exit

	workerClient := jobqueue.NewClient(cfg, postgresStore)

	// errgroup ctx: Ctrl+C cancels both roles, and a fatal error in one role
	// cancels the other.
	g, ctx := errgroup.WithContext(ctx)

	if cfg.Worker.Enabled {
		if err := workerClient.Start(ctx); err != nil {
			return err
		}
		log.Info().Msg("worker listening for jobs")
	}

	if cfg.WebServer.Enabled {
		apiServer, err := server.NewServer(
			cfg,
			postgresStore,
			workerClient,
		)
		if err != nil {
			return err
		}

		log.Info().Msgf("Platinum server listening on %s:%d", cfg.WebServer.Host, cfg.WebServer.Port)
		g.Go(func() error {
			return apiServer.ListenAndServe(ctx, cm)
		})
	}

	<-ctx.Done()    // wait for Ctrl+C or a fatal error in a role
	return g.Wait() // surface server error / graceful-shutdown result
}
