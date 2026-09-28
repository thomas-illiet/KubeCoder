package command

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/thomas-illiet/KubeCoder/backend/internal/agents"
	adminorganizationapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/admin/organizations"
	agentapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/agents"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/health"
	organizationapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/organizations"
	repositoryapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/repositories"
	"github.com/thomas-illiet/KubeCoder/backend/internal/api/server"
	userapi "github.com/thomas-illiet/KubeCoder/backend/internal/api/users"
	"github.com/thomas-illiet/KubeCoder/backend/internal/auth"
	"github.com/thomas-illiet/KubeCoder/backend/internal/config"
	"github.com/thomas-illiet/KubeCoder/backend/internal/database"
	"github.com/thomas-illiet/KubeCoder/backend/internal/logging"
	"github.com/thomas-illiet/KubeCoder/backend/internal/migration"
	"github.com/thomas-illiet/KubeCoder/backend/internal/organizations"
	"github.com/thomas-illiet/KubeCoder/backend/internal/repositories"
	"github.com/thomas-illiet/KubeCoder/backend/internal/sshkeys"
	"github.com/thomas-illiet/KubeCoder/backend/internal/users"
)

type application struct {
	version    string
	commit     string
	date       string
	configFile string
}

// New constructs the root KubeCoder command and all supported subcommands.
func New(version, commit, date string) *cobra.Command {
	app := &application{version: version, commit: commit, date: date}
	root := &cobra.Command{Use: "kubecoder", Short: "KubeCoder control-plane API", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringVar(&app.configFile, "config", "", "path to a YAML configuration file")
	root.AddCommand(app.serveCommand(), app.migrateCommand(), app.versionCommand())
	return root
}

// load resolves application configuration and initializes structured logging.
func (a *application) load(command *cobra.Command) (config.Config, *slog.Logger, error) {
	cfg, err := config.Load(a.configFile, command.Flags())
	if err != nil {
		return config.Config{}, nil, err
	}
	logger, err := logging.New(cfg.Log, os.Stdout)
	if err != nil {
		return config.Config{}, nil, err
	}
	return cfg, logger, nil
}

// addConfigFlags adds common configuration overrides to a command.
func addConfigFlags(command *cobra.Command) {
	command.Flags().String("http.address", "", "HTTP listen address")
	command.Flags().StringSlice("http.allowed_origins", nil, "allowed browser origins")
	command.Flags().String("database.dsn", "", "PostgreSQL connection string")
	command.Flags().String("oidc.issuer", "", "OIDC issuer URL")
	command.Flags().String("oidc.discovery_url", "", "optional internal OIDC discovery URL")
	command.Flags().String("oidc.audience", "", "expected access-token audience")
	command.Flags().String("log.level", "", "log level")
	command.Flags().String("log.format", "", "log format: json or text")
}

// serveCommand builds the command that runs the HTTP API.
func (a *application) serveCommand() *cobra.Command {
	command := &cobra.Command{
		Use: "serve", Short: "run the HTTP API",
		RunE: func(command *cobra.Command, _ []string) error {
			cfg, logger, err := a.load(command)
			if err != nil {
				return err
			}
			ctx, stop := signal.NotifyContext(command.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			db, err := database.Open(ctx, cfg.Database)
			if err != nil {
				return err
			}
			defer db.SQL.Close()
			verifier, err := auth.New(ctx, cfg.OIDC)
			if err != nil {
				return err
			}
			masterKey, err := cfg.OrganizationSSHKeyEncryptionKey()
			if err != nil {
				return err
			}
			sshKeyService, err := sshkeys.NewService(db.GORM, masterKey)
			if err != nil {
				return err
			}
			userService := users.NewService(users.NewRepository(db.GORM))
			userHandler := userapi.NewHandler(logger, verifier, userService)
			organizationRepository := organizations.NewRepository(db.GORM)
			organizationService := organizations.NewService(organizationRepository, sshKeyService)
			organizationHandler := organizationapi.NewHandler(logger, verifier, userService, organizationService, sshKeyService)
			adminOrganizationHandler := adminorganizationapi.NewHandler(logger, verifier, userService, organizationService)
			agentService := agents.NewService(agents.NewRepository(db.GORM))
			agentHandler := agentapi.NewHandler(logger, verifier, userService, organizationService, agentService)
			repositoryService := repositories.NewService(repositories.NewRepository(db.GORM), organizationRepository)
			repositoryHandler := repositoryapi.NewHandler(logger, verifier, userService, repositoryService)
			healthHandler := health.NewHandler(db.SQL, health.SchemaReady(db.SQL))
			handler := server.New(logger, healthHandler, userHandler, organizationHandler, adminOrganizationHandler, agentHandler, repositoryHandler, cfg.HTTP.AllowedOrigins)
			server := &http.Server{
				Addr: cfg.HTTP.Address, Handler: handler, ReadTimeout: cfg.HTTP.ReadTimeout,
				WriteTimeout: cfg.HTTP.WriteTimeout, IdleTimeout: cfg.HTTP.IdleTimeout,
			}
			errorsChannel := make(chan error, 1)
			go func() {
				logger.Info("HTTP server starting", "address", cfg.HTTP.Address, "version", a.version)
				errorsChannel <- server.ListenAndServe()
			}()
			select {
			case err := <-errorsChannel:
				if !errors.Is(err, http.ErrServerClosed) {
					return fmt.Errorf("serve HTTP: %w", err)
				}
				return nil
			case <-ctx.Done():
				shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
				defer cancel()
				logger.Info("HTTP server stopping")
				return server.Shutdown(shutdownCtx)
			}
		},
	}
	addConfigFlags(command)
	return command
}

// migrateCommand builds the parent command for database migrations.
func (a *application) migrateCommand() *cobra.Command {
	parent := &cobra.Command{Use: "migrate", Short: "manage database migrations"}
	parent.AddCommand(
		(&migrationCommand{application: a, direction: "up"}).command(),
		(&migrationCommand{application: a, direction: "down"}).command(),
		a.migrationVersionCommand(), a.migrationWaitCommand(),
	)
	return parent
}

type migrationCommand struct {
	application *application
	direction   string
}

// command builds one directional migration command.
func (m *migrationCommand) command() *cobra.Command {
	use := m.direction
	if m.direction == "down" {
		use += " [steps]"
	}
	command := &cobra.Command{Use: use, Args: cobra.MaximumNArgs(1), RunE: func(command *cobra.Command, args []string) error {
		cfg, logger, err := m.application.load(command)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(command.Context(), 2*time.Minute)
		defer cancel()
		db, err := openWithRetry(ctx, cfg.Database, logger)
		if err != nil {
			return err
		}
		defer db.SQL.Close()
		if m.direction == "up" {
			return migration.Up(db.SQL)
		}
		steps := 1
		if len(args) == 1 {
			steps, err = strconv.Atoi(args[0])
			if err != nil || steps < 1 {
				return errors.New("steps must be a positive integer")
			}
		}
		return migration.Down(db.SQL, steps)
	}}
	addConfigFlags(command)
	return command
}

// migrationVersionCommand builds the command that reports schema state.
func (a *application) migrationVersionCommand() *cobra.Command {
	command := &cobra.Command{Use: "version", RunE: func(command *cobra.Command, _ []string) error {
		cfg, _, err := a.load(command)
		if err != nil {
			return err
		}
		db, err := database.Open(command.Context(), cfg.Database)
		if err != nil {
			return err
		}
		defer db.SQL.Close()
		status, err := migration.Current(db.SQL)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(command.OutOrStdout(), "%d dirty=%t\n", status.Version, status.Dirty)
		return err
	}}
	addConfigFlags(command)
	return command
}

// migrationWaitCommand builds the command used by Kubernetes init containers.
func (a *application) migrationWaitCommand() *cobra.Command {
	var expected uint
	var timeout time.Duration
	command := &cobra.Command{Use: "wait", RunE: func(command *cobra.Command, _ []string) error {
		cfg, logger, err := a.load(command)
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(command.Context(), timeout)
		defer cancel()
		db, err := openWithRetry(ctx, cfg.Database, logger)
		if err != nil {
			return err
		}
		defer db.SQL.Close()
		return migration.Wait(ctx, db.SQL, expected, time.Second)
	}}
	command.Flags().UintVar(&expected, "version", migration.LatestVersion, "required schema version")
	command.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "maximum wait duration")
	addConfigFlags(command)
	return command
}

// openWithRetry waits for PostgreSQL to become reachable until context cancellation.
func openWithRetry(ctx context.Context, cfg config.DatabaseConfig, logger *slog.Logger) (*database.DB, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		db, err := database.Open(ctx, cfg)
		if err == nil {
			return db, nil
		}
		logger.Warn("database is not ready", "error", err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait for database: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

// versionCommand builds the command that prints build metadata.
func (a *application) versionCommand() *cobra.Command {
	return &cobra.Command{Use: "version", RunE: func(command *cobra.Command, _ []string) error {
		_, err := fmt.Fprintf(command.OutOrStdout(), "kubecoder %s commit=%s built=%s\n", a.version, a.commit, a.date)
		return err
	}}
}
