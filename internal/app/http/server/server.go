package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/proxeter/errors-channel"
	"go.elastic.co/apm/module/apmhttp"
	"go.uber.org/zap"

	"github.com/denisandreenko/testservice/internal/app/config"
)

type (
	Server struct {
		ctx context.Context

		logger *zap.Logger
		config *config.AppConfig
	}

	panicHandler struct {
		logger *zap.Logger

		handler http.Handler
	}
)

func (h panicHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	urlQuery := req.URL.Query()
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.logger.Error("error while read body for recover", zap.Error(err))

		w.WriteHeader(http.StatusNoContent)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			h.logger.Error(
				"http request recovered",
				zap.Any("error", err),
				zap.Any("query", urlQuery),
				zap.ByteString("body", data),
			)
		}
	}()

	req.Body = ioutil.NopCloser(bytes.NewReader(data))
	h.handler.ServeHTTP(w, req)
}

func NewServer(
	ctx context.Context,
	logger *zap.Logger,
	config *config.AppConfig,
) <-chan error {
	return errch.Register(func() error {
		return (&Server{
			ctx: ctx,

			logger: logger,
			config: config,
		}).Start(ctx)
	})
}

func (as *Server) Start(ctx context.Context) error {
	server := http.Server{
		Addr:    ":" + strconv.Itoa(as.config.HTTP.Port),
		Handler: as.handlers(),
	}

	hf := func() error {
		return server.ListenAndServe()
	}

	as.logger.Info("http server is running", zap.String("host", as.config.HTTP.Host), zap.Int("port", as.config.HTTP.Port))
	select {
	case err := <-errch.Register(hf):
		as.logger.Info("Shutdown http server", zap.String("by", "error"), zap.Error(err))
		return server.Shutdown(ctx)
	case <-ctx.Done():
		as.logger.Info("Shutdown http server", zap.String("by", "context.Done"))
		return server.Shutdown(ctx)
	}
}

func (as *Server) ph(handler http.HandlerFunc) http.Handler {
	return apmhttp.Wrap(panicHandler{
		logger:  as.logger,
		handler: handler,
	})
}

func (as *Server) handlers() *httprouter.Router {
	return func() *httprouter.Router {
		router := httprouter.New()

		router.Handler(http.MethodGet, "/", as.ph(hw))

		router.HandlerFunc(http.MethodGet, "/debug/pprof/", pprof.Index)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/cmdline", pprof.Cmdline)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/profile", pprof.Profile)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/symbol", pprof.Symbol)
		router.HandlerFunc(http.MethodGet, "/debug/pprof/trace", pprof.Trace)
		router.Handler(http.MethodGet, "/debug/pprof/heap", pprof.Handler("heap"))
		router.Handler(http.MethodGet, "/debug/pprof/block", pprof.Handler("block"))
		router.Handler(http.MethodGet, "/debug/pprof/mutex", pprof.Handler("mutex"))
		router.Handler(http.MethodGet, "/debug/pprof/allocs", pprof.Handler("allocs"))
		router.Handler(http.MethodGet, "/debug/pprof/goroutine", pprof.Handler("goroutine"))
		router.Handler(http.MethodGet, "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

		router.Handler(http.MethodGet, "/metrics", promhttp.Handler())
		router.HandlerFunc(http.MethodGet, "/check", check)

		return router
	}()
}

func hw(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello World"))
}

func check(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
