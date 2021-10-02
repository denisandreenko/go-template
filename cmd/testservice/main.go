package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/chapsuk/grace"
	_ "go.uber.org/automaxprocs" // optimization for k8s
	"go.uber.org/zap"

	application "github.com/denisandreenko/testservice/internal/app"
	"github.com/denisandreenko/testservice/internal/app/config"
	ll "github.com/denisandreenko/testservice/internal/app/logger"
	"github.com/denisandreenko/testservice/internal/pkg/env"
)

const (
	Name = "testservice"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	ctx := grace.ShutdownContext(context.Background())

	logger, err := ll.New(
		Name,
		os.Getenv("VERSION"),
		os.Getenv("ENV"),
		os.Getenv("LOG_LEVEL"),
	)
	if err != nil {
		log.Fatal("error while init logger", zap.Error(err))
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("recovered error", zap.Any("description", r))
		}
	}()

	appConfig, err := config.NewAppConfig(os.Getenv("CONFIG_PATH"))
	if err != nil {
		logger.Fatal("error while init config", zap.Error(err))
	}

	app := application.New(
		Name,
		os.Getenv("VERSION"),
		os.Getenv("ENV"),
		appConfig,
		logger,
	)

	ctx = context.WithValue(ctx, env.Name, app.Name)
	ctx = context.WithValue(ctx, env.Version, app.Version)
	ctx = context.WithValue(ctx, env.Environment, app.Environment)
	ctx = context.WithValue(ctx, env.Tags, []string{app.Version, app.Environment})

	app.Run(ctx)
	app.Shutdown()
}
