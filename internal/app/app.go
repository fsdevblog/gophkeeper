package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsdevblog/gophkeeper/internal/config"
	"github.com/fsdevblog/gophkeeper/internal/storage/services"
	apphttp "github.com/fsdevblog/gophkeeper/internal/transport/http"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultShutdownTimeout  = 5 * time.Second
	defaultMigrationTimeout = 60 * time.Second
)

type Options struct {
	ShutdownTimeout  time.Duration
	MigrationTimeout time.Duration
}

type App struct {
	config            *config.Config
	l                 *zap.Logger
	serviceCollection *services.Collection

	shutdownTimeout time.Duration
}

func Must(a *App, err error) *App {
	if err != nil {
		panic(err)
	}
	return a
}

func New(conf *config.Config, l *zap.Logger, opts ...func(*Options)) (*App, error) {
	options := Options{
		ShutdownTimeout:  defaultShutdownTimeout,
		MigrationTimeout: defaultMigrationTimeout,
	}
	for _, opt := range opts {
		opt(&options)
	}
	initServiceCtx, cancel := context.WithTimeout(context.Background(), options.MigrationTimeout)
	defer cancel()

	collection, errCollection := services.NewCollection(initServiceCtx, conf)
	if errCollection != nil {
		return nil, errCollection
	}
	return &App{
		config:            conf,
		l:                 l,
		serviceCollection: collection,
		shutdownTimeout:   options.ShutdownTimeout,
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := a.startHTTPServer(ctx)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("app run: %w", err)
	}
	return nil
}

func (a *App) startHTTPServer(ctx context.Context) error {
	router, errRouter := apphttp.New(apphttp.InitArgs{
		JWTSecret:   []byte(a.config.JWTSecret),
		AuthService: a.serviceCollection.AuthService,
	})
	if errRouter != nil {
		return fmt.Errorf("start HTTP server: %w", errRouter)
	}

	httpSrv := &http.Server{
		Addr:    a.config.HTTPServerAddr,
		Handler: router,
	}
	lis, errLis := net.Listen("tcp", a.config.HTTPServerAddr)
	if errLis != nil {
		return fmt.Errorf("start HTTP server: %w", errLis)
	}

	var serverError error
	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, a.shutdownTimeout)
		defer shutdownCancel()
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			serverError = errors.Join(serverError, fmt.Errorf("shutdown HTTP server: %w", err))
		}
	}()

	serverError = httpSrv.Serve(lis)
	return serverError
}
