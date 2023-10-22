package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"scan_project/configuration"
	"scan_project/internal/dao"
	"scan_project/internal/kube"
	"syscall"
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
	}
	scansDao := dao.NewScansDao(logrus.NewEntry(logger).WithField("app", "scansDAO"))

	// Start KubeScanner
	kubeScanner := kube.NewKubeScanner(
		postgresDB,
		&scansDao,
		config.System.Kubernetes.Timeout,
		logrus.NewEntry(logger).WithField("app", "KubeScanner"),
	)
	go kubeScanner.Start(config.ScanDelay)
	defer kubeScanner.Shutdown()

	exitChl := make(chan os.Signal, 1)
	signal.Notify(exitChl, syscall.SIGINT, syscall.SIGTERM)
	<-exitChl
}
