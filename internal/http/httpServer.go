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
	logger  *logrus.Entry
	storage kube.StorageI
}

func NewHttpServer(storage kube.StorageI, loggerEntry *logrus.Entry, cfg *configuration.Config) *http.Server {
	httpServer := HttpServer{
		logger:  loggerEntry,
		storage: storage,
	}
	r := mux.NewRouter()
	r.Use(httpServer.loggingMiddleware)
	r.HandleFunc("/cluster/{cluster}/namespace/{namespace}/jobs-scans", httpServer.getJobsScans).Methods("GET")
	r.HandleFunc("/cluster/{cluster}/namespace/{namespace}/services-scans", httpServer.getServicesScans).Methods("GET")
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.System.Http.Port),
		Handler:      r,
		ReadTimeout:  time.Second * time.Duration(cfg.System.Http.Timeout),
		WriteTimeout: time.Second * time.Duration(cfg.System.Http.Timeout),
	}
}
