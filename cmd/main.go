package main

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	http "net/http"
	"os"
	"os/signal"
	"scan_project/configuration"
	"scan_project/internal/dao"
	"scan_project/internal/httpServer"
	"scan_project/internal/kube"
	"syscall"
	"time"
)

const (
	pathToConfig    = "/home/reserv/GolandProjects/LogScan/config.json"
	DefaultLogLevel = logrus.InfoLevel
)

// TODO: исправить проброс ошибок из БД и обрабатывать их в HTTP (переписать хэндлер ошибок)
// TODO: Написать Swagger-файл

// TODO: остается проблема, когда сканы старых добавленных конфигов не стираются из памяти
// TODO: русский PostgreSQL... Надо доработать regexp на такой случай (???)

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
	scansDao := dao.NewScansDao(logrus.NewEntry(logger).WithField("app", "scans-in-memory"))
	storage := dao.NewStorage(postgresDB, &scansDao)

	// Start KubeScanner
	kubeScanner := kube.NewKubeScanner(
		storage,
		config.System.Kubernetes.Timeout,
		logrus.NewEntry(logger).WithField("app", "kube-scanner"),
	)
	go kubeScanner.Start(config.ScanDelay)
	defer kubeScanner.Shutdown()

	// Start httpServer.server
	server := httpServer.NewHttpServer(storage, logger.WithField("app", "httpServer-server"), config)
	go func() {
		err := server.ListenAndServe()
		if err == http.ErrServerClosed {
			logger.Info("Http server was closed")
		} else {
			logger.
				WithField("error", err).
				Error("Failed to start the HTTP Server")
		}
	}()
	defer func(server *http.Server) {
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
