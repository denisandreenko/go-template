package app

import (
	"context"

	"go.uber.org/zap"

	"github.com/denisandreenko/testservice/internal/app/config"
	"github.com/denisandreenko/testservice/internal/app/grpc/server"
	http "github.com/denisandreenko/testservice/internal/app/http/server"
)

type (
	Application struct {
		config *config.AppConfig
		logger *zap.Logger

		Name        string
		Version     string
		Environment string
	}
)

func New(
	name, version, environment string,
	config *config.AppConfig,
	logger *zap.Logger,
) *Application {
	return &Application{
		config: config,
		logger: logger,

		Name:        name,
		Version:     version,
		Environment: environment,
	}
}

func (app *Application) Run(ctx context.Context) {
	grpcServerErrCh := grpc.NewServer(ctx, app.logger, app.config)
	httpServerErrCh := http.NewServer(ctx, app.logger, app.config)

	select {
	case <-grpcServerErrCh:
		<-httpServerErrCh
	case <-httpServerErrCh:
		<-grpcServerErrCh
	}
}

func (app *Application) Shutdown() {
	_ = app.logger.Sync()
}
