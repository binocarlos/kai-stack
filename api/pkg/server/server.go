package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"

	"github.com/binocarlos/kai-stack/api/pkg/config"
	"github.com/binocarlos/kai-stack/api/pkg/jobqueue"
	"github.com/binocarlos/kai-stack/api/pkg/store"
	"github.com/binocarlos/kai-stack/api/pkg/system"

	_ "net/http/pprof" // enable profiling
)

type StackAPIServer struct {
	app      *fiber.App
	router   fiber.Router
	cfg      *config.Config
	store    *store.PostgresStore
	jobqueue *jobqueue.Client
}

func NewServer(
	cfg *config.Config,
	store *store.PostgresStore,
	workerClient *jobqueue.Client,
) (*StackAPIServer, error) {
	if cfg.WebServer.Host == "" {
		return nil, fmt.Errorf("server host is required")
	}

	if cfg.WebServer.Port == 0 {
		return nil, fmt.Errorf("server port is required")
	}

	app := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024 * 1024, // 100GB
	})

	app.Use(logger.New(logger.Config{
		Next: func(c fiber.Ctx) bool {
			// Check if the 'nolog' query parameter exists
			if c.Query("nolog") != "" {
				// Skip logging if 'nolog' is present
				return true
			}
			// Continue logging otherwise
			return false
		},
	}))

	server := &StackAPIServer{
		app:      app,
		router:   app.Group(cfg.WebServer.APIPath),
		cfg:      cfg,
		store:    store,
		jobqueue: workerClient,
	}

	server.RegisterUserRoutes()
	server.RegisterComicRoutes()

	return server, nil
}

func (apiServer *StackAPIServer) ListenAndServe(ctx context.Context, _ *system.CleanupManager) error {
	addr := fmt.Sprintf("%s:%d", apiServer.cfg.WebServer.Host, apiServer.cfg.WebServer.Port)
	return apiServer.app.Listen(addr)
}

func getRequestData[TBodyData any](c fiber.Ctx) (*TBodyData, error) {
	var bodyData TBodyData
	if err := c.Bind().Body(&bodyData); err != nil {
		return nil, err
	}

	return &bodyData, nil
}

// RegisterComicRoutes registers all comic-related routes
func (apiServer *StackAPIServer) RegisterComicRoutes() {
	comicRouter := NewComicRouter(apiServer, apiServer.store.Comics())
	comicRouter.RegisterRoutes(apiServer.router)
}
