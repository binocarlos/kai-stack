package goapi

import (
	"context"
	"os"
	"os/signal"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/binocarlos/kai-stack/api/pkg/jobqueue"
	"github.com/binocarlos/kai-stack/api/pkg/store"
	"github.com/binocarlos/kai-stack/api/pkg/system"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newWorkerCmd() *cobra.Command {
	workerConfig, err := newConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create options")
	}

	envHelpText := generateEnvHelpText(workerConfig, "")

	serveCmd := &cobra.Command{
		Use:     "worker",
		Short:   "Start the forum worker.",
		Long:    "Start the forum worker.",
		Example: "TBD",
		RunE: func(cmd *cobra.Command, _ []string) error {
			err := worker(cmd, workerConfig)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to run server")
			}
			return nil
		},
	}

	serveCmd.Long += "\n\nEnvironment Variables:\n\n" + envHelpText

	return serveCmd
}

func worker(cmd *cobra.Command, cfg *config.Config) error {
	system.SetupLogging()

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

	workerClient, err := jobqueue.NewClient(ctx, cfg, postgresStore)
	if err != nil {
		return err
	}

	err = workerClient.Start()
	if err != nil {
		return err
	}

	log.Info().Msgf("Platinum worker listening for jobs")

	<-ctx.Done()
	return nil
}
