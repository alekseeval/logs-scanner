package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"scan_project/configuration"
	"scan_project/internal/dal"
	"scan_project/internal/httpServer"
	"scan_project/internal/kube"
	"syscall"
	"time"
)

const (
	pathToConfig               = "/etc/scanner/config.json"
	DefaultLogLevel            = logrus.InfoLevel
	HttpServerShutdownTimeout  = 5 * time.Second
	KubeScannerShutdownTimeout = 20 * time.Second
)

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	config, err := configuration.ReadConfig(pathToConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to read config file %s", pathToConfig))
	}
	logger.
		WithField("config", config).
		Info("Config was parsed")
	lvl, err := logrus.ParseLevel(config.Logger.Level)
	if err != nil {
		lvl = DefaultLogLevel
		logger.Error("failed to parse log level, will be used " + DefaultLogLevel.String() + " as default")
	}
	logger.SetLevel(lvl)
	logger.Debugf("set log level to %s", lvl)

	// Init DAO
	postgresDB, err := dal.NewPostgresDB(config, logrus.NewEntry(logger).WithField("app", "postgresql"))
	if err != nil {
		logger.
			WithField("error", err).
			Error("Failed to init postgres DB")
		return
	}
	scansDao := dal.NewScansDao(logrus.NewEntry(logger).WithField("app", "scans-in-memory"))
	storage := dal.NewStorage(postgresDB, &scansDao)

	// Start KubeScanner
	kubeScanner := kube.NewKubeScanner(
		storage,
		config,
		logrus.NewEntry(logger).WithField("app", "kube-scanner"),
	)
	go func() {
		err = kubeScanner.Start(config.ScanDelay)
		if err != nil {
			logger.
				WithField("error", err).
				Error("failed to start kube-scanner")
		}
		return
	}()
	defer func() {
		ksCtx, ctxCancel := context.WithTimeout(context.Background(), KubeScannerShutdownTimeout)
		defer ctxCancel()
		err = kubeScanner.Shutdown(ksCtx)
		if err != nil {
			logger.WithField("error", err).Error("Failed to gracefully shutdown kube-scanner")
		}
	}()

	// Start httpServer.server
	server := httpServer.NewHttpServer(config, storage, logger.WithField("app", "httpServer-server"))
	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("Http server was closed")
		} else {
			logger.
				WithField("error", err).
				Error("Failed to start the HTTP Server")
		}
	}()
	defer func(server *http.Server) {
		ctx, ctxCancel := context.WithTimeout(context.Background(), HttpServerShutdownTimeout)
		defer ctxCancel()
		err := server.Shutdown(ctx)
		if err != nil {
			logger.
				WithField("error", err).
				Error("Failed to shutdown the HTTP Server gracefully")
		}
	}(server)

	exitChl := make(chan os.Signal, 1)
	signal.Notify(exitChl, syscall.SIGINT, syscall.SIGTERM)
	<-exitChl
}
