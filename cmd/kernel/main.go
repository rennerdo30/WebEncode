package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rennerdo30/webencode/internal/api/handlers"
	apimiddleware "github.com/rennerdo30/webencode/internal/api/middleware"
	"github.com/rennerdo30/webencode/internal/cleanup"
	"github.com/rennerdo30/webencode/internal/events"
	"github.com/rennerdo30/webencode/internal/live"
	"github.com/rennerdo30/webencode/internal/metrics"
	"github.com/rennerdo30/webencode/internal/orchestrator"
	"github.com/rennerdo30/webencode/internal/plugin_manager"
	"github.com/rennerdo30/webencode/internal/webhooks"
	"github.com/rennerdo30/webencode/internal/workers"
	"github.com/rennerdo30/webencode/pkg/bus"
	"github.com/rennerdo30/webencode/pkg/config"
	"github.com/rennerdo30/webencode/pkg/db/migrate"
	"github.com/rennerdo30/webencode/pkg/db/store"
	"github.com/rennerdo30/webencode/pkg/logger"
)

func main() {
	l := logger.New("kernel")

	if err := run(l); err != nil {
		l.Error("Kernel crashed", "error", err)
		os.Exit(1)
	}
}

func run(l *logger.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Config
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	l.Info("Config loaded", "port", cfg.Port)

	// Determine migrations path
	migrationsPath := "./migrations"
	if _, err := os.Stat("pkg/db/migrations"); err == nil {
		migrationsPath = "pkg/db/migrations"
	}
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		migrationsPath = envPath
	}

	// Run Migrations
	if err := migrate.Run(cfg.DatabaseURL, migrationsPath); err != nil {
		l.Error("Failed to run migrations", "error", err)
		return err
	}
	l.Info("Database migrations applied")

	// 2. Database (PGX Pool)
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		return err
	}
	l.Info("Database connected")

	// 3. NATS Bus
	b, err := bus.Connect(cfg.NatsURL, l)
	if err != nil {
		return err
	}
	defer b.Close()

	if err := b.InitStreams(ctx); err != nil {
		return err
	}
	l.Info("Messaging bus initialized")

	// 4. Plugin Manager
	pm := plugin_manager.New(l, cfg.PluginDir)
	defer pm.Shutdown()

	if err := pm.LoadAll(); err != nil {
		l.Error("Failed to load plugins (continuing)", "error", err)
	}
	l.Info("Plugin Manager started")

	// Register loaded plugins in DB
	querier := store.New(db)
	registerPlugins(ctx, querier, pm, l)

	// 5. HTTP Server
	r := chi.NewRouter()
	r.Use(middleware.Logger)                            // Chi middleware
	r.Use(apimiddleware.InjectDependencies(querier, l)) // Inject Logger/DB
	r.Use(apimiddleware.Recovery(querier, l))           // Custom Recovery Middleware
	r.Use(apimiddleware.CORS)                           // CORS for cross-origin requests

	// Rate limiting: 100 requests per minute, burst of 20
	rateLimiter := apimiddleware.DefaultRateLimiter()
	r.Use(apimiddleware.RateLimit(rateLimiter))

	// Handlers
	// querier is already defined above

	orch := orchestrator.New(querier, b, l)
	jobsHandler := handlers.NewJobsHandler(orch, pm, l)
	jobsHandler.Register(r)

	streamsHandler := handlers.NewStreamsHandler(querier, l)
	streamsHandler.Register(r)

	workersHandler := handlers.NewWorkersHandler(querier, l)
	workersHandler.Register(r)

	restreamsHandler := handlers.NewRestreamsHandler(querier, orch, l)
	restreamsHandler.Register(r)

	profilesHandler := handlers.NewProfilesHandler(querier, l)
	profilesHandler.Register(r)

	pluginsHandler := handlers.NewPluginsHandler(querier, pm, l)
	pluginsHandler.Register(r)

	auditHandler := handlers.NewAuditHandler(querier, l)
	auditHandler.Register(r)

	sseHandler := handlers.NewSSEHandler(b, l)
	sseHandler.Register(r)

	webhooksHandler := handlers.NewWebhooksHandler(querier, l)
	webhooksHandler.Register(r)

	errorsHandler := handlers.NewErrorsHandler(querier, l)
	errorsHandler.Register(r)

	systemHandler := handlers.NewSystemHandler(querier, l)
	systemHandler.Register(r)

	filesHandler := handlers.NewFilesHandler(pm, l)
	filesHandler.Register(r)

	notificationsHandler := handlers.NewNotificationsHandler(querier, l)
	notificationsHandler.Register(r)

	// Prometheus metrics endpoint
	r.Handle("/metrics", metrics.Handler())

	liveHandler := handlers.NewLiveHandler(l, querier)
	r.Post("/v1/live/auth", liveHandler.HandleAuth)
	r.Post("/v1/live/start", liveHandler.HandleStart)
	r.Post("/v1/live/stop", liveHandler.HandleStop)

	// Live Monitor Service
	liveMonitor := live.NewMonitorService(querier, b, pm, l)
	liveMonitor.Start()
	defer liveMonitor.Stop()

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		l.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.Error("Server error", "error", err)
		}
	}()

	// 6. Webhooks Manager
	wm := webhooks.New(b, querier, l)
	go wm.Start(ctx)

	// 7. Worker Heartbeat Listener
	hl := workers.NewHeartbeatListener(b, querier, l)
	go hl.Start(ctx)

	// 8. Job Event Listener
	el := events.NewEventListener(b, querier, orch, l)
	go el.Start(ctx)

	// 9. Cleanup Service (GC for old jobs, workers, etc.)
	cs := cleanup.New(querier, l, cleanup.DefaultConfig())
	go cs.Start(ctx)

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	return srv.Shutdown(shutdownCtx)
}

func registerPlugins(ctx context.Context, db store.Querier, pm *plugin_manager.Manager, l *logger.Logger) {
	// Helper to register a list of plugins
	register := func(pluginType string, plugins map[string]interface{}) {
		for id := range plugins {
			l.Info("Registering plugin", "id", id, "type", pluginType)
			_, err := db.RegisterPluginConfig(ctx, store.RegisterPluginConfigParams{
				ID:         id,
				PluginType: pluginType,
				ConfigJson: []byte("{}"),
				IsEnabled:  pgtype.Bool{Bool: true, Valid: true}, // Default to enabled
				Priority:   pgtype.Int4{Int32: 0, Valid: true},
			})
			if err != nil {
				l.Warn("Failed to register plugin", "id", id, "error", err)
			}
		}
	}

	register("auth", pm.Auth)
	register("storage", pm.Storage)
	register("encoder", pm.Encoder)
	register("live", pm.Live)
	register("publisher", pm.Publisher)
}
