package http

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"scan_project/configuration"
	"scan_project/internal/kube"
	"time"
)

// TODO: логировать запросы
// TODO: унифицировать обработку ошибок записи json
// TODO: дописать хендлеры на остальные запросы

type HttpServer struct {
	logger      *logrus.Entry
	clustersDAO kube.ClusterDAOI
	scansDAO    kube.ScansDAOI
}

func NewHttpServer(clusterDAO kube.ClusterDAOI, scansDAO kube.ScansDAOI, loggerEntry *logrus.Entry, cfg *configuration.Config) *http.Server {
	httpServer := HttpServer{
		logger:      loggerEntry,
		clustersDAO: clusterDAO,
		scansDAO:    scansDAO,
	}
	r := mux.NewRouter()
	r.HandleFunc("/cluster/{cluster}/namespace/{namespace}/jobs-scans", httpServer.getJobsScans).Methods("GET")
	r.HandleFunc("/cluster/{cluster}/namespace/{namespace}/services-scans", httpServer.getServicesScans).Methods("GET")
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.System.Http.Port),
		Handler:      r,
		ReadTimeout:  time.Second * time.Duration(cfg.System.Http.Timeout),
		WriteTimeout: time.Second * time.Duration(cfg.System.Http.Timeout),
	}
}
