package main

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/uptrace/bun"
	"go.uber.org/fx"

	"github.com/hackerchai/memento/internal/api"
	"github.com/hackerchai/memento/internal/auth"
	"github.com/hackerchai/memento/internal/config"
	"github.com/hackerchai/memento/internal/entity"
	"github.com/hackerchai/memento/internal/middleware"
	"github.com/hackerchai/memento/internal/repository"
	"github.com/hackerchai/memento/internal/router"
	"github.com/hackerchai/memento/internal/service"
	"github.com/hackerchai/memento/internal/sse"
	"github.com/hackerchai/memento/internal/storage"
	orig_db "github.com/hackerchai/memento/pkg/db"
	"github.com/hackerchai/memento/pkg/xlog"
)

func main() {
	fx.New(
		// Provide core components
		fx.Provide(
			// Configuration
			func() (*config.Config, error) {
				cfg, err := config.LoadConfig(".")
				if err != nil {
					return nil, err
				}
				return &cfg, nil
			},
			// Logger (Original Provider)
			provideXlogLogger,
			// Database Connection
			func(cfg *config.Config) (*bun.DB, error) {
				db, err := orig_db.NewDBConnection(cfg.Database.Driver, cfg.Database.Source, cfg.App.DebugMode)
				if err != nil {
					return nil, err
				}
				// Register all models in a single call, adjust order for M2M
				db.RegisterModel(
					(*entity.User)(nil),
					(*entity.Category)(nil),
					(*entity.ArticleTag)(nil), // Register join table first (or early)
					(*entity.Tag)(nil),        // Then register models involved in M2M
					(*entity.Article)(nil),    // Then register models involved in M2M
				)

				return db, nil
			},
			// Fiber App (depends on logger)
			provideFiberApp,
		),
		// Include modules
		auth.Module,
		repository.Module,
		storage.Module,
		service.Module,
		api.Module,
		sse.Module,
		router.Module,

		// Invoke setup functions
		fx.Invoke(
			registerAppHooks,
			runDBMigrations,
		),
	).Run()
}

// provideXlogLogger creates the xlog logger based on configuration.
func provideXlogLogger(cfg *config.Config) (*xlog.Logger, error) {
	var output io.Writer
	if !cfg.Log.Enable {
		output = io.Discard
		// Assume trace setting doesn't matter if disabled
		return xlog.NewLogger(output, zerolog.Disabled, "", nil), nil
	}

	if cfg.Log.File != "" {
		output = xlog.NewRotateWriter(
			xlog.WithFilename(cfg.Log.File),
			xlog.WithMaxSize(cfg.Log.MaxSize),
			xlog.WithMaxAge(cfg.Log.MaxAge),
			xlog.WithMaxBackups(cfg.Log.MaxBackups),
			xlog.WithCompress(cfg.Log.Compress),
		)
	} else {
		if cfg.Log.Beautify {
			output = xlog.NewConsoleWriter(os.Stdout, time.RFC3339, false)
		} else {
			output = os.Stdout
		}
	}

	level, err := zerolog.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	if cfg.App.DebugMode {
		level = zerolog.DebugLevel
	}

	logger := xlog.NewLogger(output, level, "", nil)
	if cfg.Log.Trace {
		logger.SetTrace(true)
	}

	return logger, nil
}

// provideFiberApp creates and configures the Fiber application.
func provideFiberApp(lc fx.Lifecycle, logger *xlog.Logger, config *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			log := xlog.Ctx(ctx.UserContext())
			log.ErrorX(ctx.UserContext()).Err(err).Msg("Request Error")

			return ctx.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	// Core middlewares
	app.Use(recover.New())
	app.Use(cors.New())

	// Use our custom xlog middleware for request ID and logging
	app.Use(middleware.XlogMiddleware(logger, config))

	return app
}

// runDBMigrations runs database migrations.
func runDBMigrations(lc fx.Lifecycle, cfg *config.Config, logger *xlog.Logger) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return RunMigrations(cfg.Database, logger)
		},
	})
}

// registerAppHooks sets up lifecycle hooks.
type AppHooksParams struct {
	fx.In
	LC  fx.Lifecycle
	DB  *bun.DB
	App *fiber.App
	Cfg *config.Config
	Log *xlog.Logger
}

func registerAppHooks(p AppHooksParams) {
	p.LC.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.Log.InfoX().Msg("Executing OnStart hook...")
			p.Log.InfoX().Msg("Database connection established (migrations run separately).")
			go func() {
				p.Log.InfoX().Msgf("Starting server on %s", p.Cfg.Server.Address)
				if err := p.App.Listen(p.Cfg.Server.Address); err != nil {
					p.Log.ErrorX(ctx).Err(err).Msg("Failed to start server")
				}
			}()
			p.Log.InfoX().Msg("OnStart hook completed.")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			p.Log.InfoX().Msg("Executing OnStop hook...")
			p.Log.InfoX().Msg("Shutting down server...")
			if err := p.App.Shutdown(); err != nil {
				p.Log.ErrorX().Err(err).Msg("Error shutting down server")
			}
			if err := p.DB.Close(); err != nil {
				p.Log.ErrorX().Err(err).Msg("Error closing database connection")
				return err
			}
			p.Log.InfoX().Msg("Server gracefully stopped.")
			return nil
		},
	})
}
