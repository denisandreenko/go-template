package grpc

import (
	"context"
	"net"
	"strconv"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	_ "github.com/jnewmano/grpc-json-proxy/codec" // GRPC Proxy
	errch "github.com/proxeter/errors-channel"
	"go.elastic.co/apm/module/apmgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	testservice "github.com/denisandreenko/testservice/api"
	"github.com/denisandreenko/testservice/internal/app/config"
)

type (
	GRPC struct {
		logger *zap.Logger
		config *config.AppConfig
	}
)

func NewServer(
	ctx context.Context,
	logger *zap.Logger,
	config *config.AppConfig,
) <-chan error {
	return errch.Register(func() error {
		return (&GRPC{
			logger: logger,
			config: config,
		}).Start(ctx)
	})
}

func (g *GRPC) Start(ctx context.Context) error {
	address := ":" + strconv.Itoa(g.config.GRPC.Port)
	conn, err := net.Listen("tcp4", address)
	if err != nil {
		g.logger.Fatal("error while listen socket for grpc service", zap.Error(err))
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_prometheus.UnaryServerInterceptor,
				apmgrpc.NewUnaryServerInterceptor(apmgrpc.WithRecovery()),
			),
		),
	)

	testservice.RegisterHelloWorldServer(server, g)
	healthpb.RegisterHealthServer(server, health.NewServer())

	grpc_prometheus.EnableHandlingTimeHistogram()
	grpc_prometheus.Register(server)

	reflection.Register(server)

	g.logger.Info("Start grpc server", zap.String("address", address))
	select {
	case err := <-errch.Register(func() error { return server.Serve(conn) }):
		return err
	case <-ctx.Done():
		g.logger.Info("Shutdown grpc server")
		server.GracefulStop()

		return ctx.Err()
	}
}

func (g *GRPC) Get(ctx context.Context, req *testservice.Request) (*testservice.Response, error) {
	return &testservice.Response{Message: "Hello World"}, nil
}
