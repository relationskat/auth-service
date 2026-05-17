package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpcctrl "github.com/relationskat/auth-service/internal/controller/grpc"
	"github.com/relationskat/auth-service/internal/cache/memcached"
	"github.com/relationskat/auth-service/internal/config"
	"github.com/relationskat/auth-service/internal/pkg/provider"
	"github.com/relationskat/auth-service/internal/repository/store"
	"github.com/relationskat/auth-service/internal/service/production"
	"github.com/relationskat/auth-service/internal/smtp"
	"github.com/relationskat/auth-service/pkg/db"
	authv1 "github.com/relationskat/auth-service/pkg/gen"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() { _ = log.Sync() }()

	cfg, err := config.New()
	if err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	ctx := context.Background()

	pool, err := db.NewPostgres(ctx, cfg)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	memClient, err := db.NewCache(cfg)
	if err != nil {
		log.Fatal("failed to create cache client", zap.Error(err))
	}

	repo := store.New(pool, log)
	tokens := memcached.New(log, memClient)

	mailer, err := smtp.New(log, *cfg)
	if err != nil {
		log.Fatal("failed to create mailer", zap.Error(err))
	}

	prov, err := provider.New(cfg)
	if err != nil {
		log.Fatal("failed to create token provider", zap.Error(err))
	}

	svc, err := production.New(log, repo, tokens, mailer, prov, cfg.HTTP.BaseURL())
	if err != nil {
		log.Fatal("failed to create service", zap.Error(err))
	}

	handler := grpcctrl.New(log, svc)

	server := grpc.NewServer()
	authv1.RegisterAuthServer(server, handler)

	addr := net.JoinHostPort(cfg.GRPC.Host, cfg.GRPC.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to listen", zap.String("addr", addr), zap.Error(err))
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("grpc server started", zap.String("addr", addr))
		errCh <- server.Serve(listener)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal("grpc server stopped", zap.Error(err))
		}
	case sig := <-stop:
		log.Info("shutting down", zap.String("signal", sig.String()))
		server.GracefulStop()
	}
}
