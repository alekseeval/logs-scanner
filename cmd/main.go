package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	http2 "net/http"
	"os"
	"os/signal"
	"scan_project/configuration"
	"scan_project/internal/dao"
	"scan_project/internal/http"
	"scan_project/internal/kube"
	"syscall"
	"time"
)

const (
	pathToConfig    = "/home/reserv/GolandProjects/LogScan/config.json"
	DefaultLogLevel = logrus.InfoLevel
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
	postgresDB, err := dao.NewPostgresDB(config, logrus.NewEntry(logger).WithField("app", "postgresql"))
	if err != nil {
		logger.
			WithField("error", err).
			Error("Failed to init postgres DB")
		return
	}
	scansDao := dao.NewScansDao(logrus.NewEntry(logger).WithField("app", "scansDAO"))
	storage := dao.NewStorage(postgresDB, &scansDao)

	// Start KubeScanner
	kubeScanner := kube.NewKubeScanner(
		storage,
		config.System.Kubernetes.Timeout,
		logrus.NewEntry(logger).WithField("app", "KubeScanner"),
	)
	go kubeScanner.Start(config.ScanDelay)
	defer kubeScanner.Shutdown()

	// Start http.server
	server := http.NewHttpServer(storage, logger.WithField("app", "http-server"), config)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.
				WithField("error", err).
				Error("Failed to start the HTTP Server")
		}
	}()
	defer func(server *http2.Server) {
		ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
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
