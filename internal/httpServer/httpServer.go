package httpServer

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"scan_project/configuration"
	"scan_project/internal/kube"
	"time"
)

type httpServer struct {
	logger  *logrus.Entry
	storage kube.StorageI
}

func NewHttpServer(cfg *configuration.Config, storage kube.StorageI, loggerEntry *logrus.Entry) *http.Server {
	httpServer := httpServer{
		logger:  loggerEntry,
		storage: storage,
	}
	r := mux.NewRouter()
	r.Use(httpServer.loggingMiddleware) // Log request
	r.Use(setResponseHeadersMiddleware) // set CORS and Content-Type headers
	// Clusters
	r.HandleFunc("/api/v1/clusters", httpServer.getAllClusters).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/clusters/{cluster}", httpServer.getCluster).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/clusters", httpServer.createCluster).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/clusters/{cluster}", httpServer.deleteCluster).Methods(http.MethodDelete)
	r.HandleFunc("/api/v1/clusters/{cluster}/namespaces", httpServer.addNamespace).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}", httpServer.deleteNamespace).Methods(http.MethodDelete)
	r.HandleFunc("/api/v1/clusters/{cluster}/config", httpServer.changeClusterConfig).Methods(http.MethodPatch)
	// Scans
	r.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/jobs-scans", httpServer.getJobsScans).Methods(http.MethodGet)
	r.HandleFunc("/api/v1/clusters/{cluster}/namespaces/{namespace}/services-scans", httpServer.getServicesScans).Methods(http.MethodGet)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.System.Http.Port),
		Handler:      r,
		ReadTimeout:  time.Second * time.Duration(cfg.System.Http.Timeout),
		WriteTimeout: time.Second * time.Duration(cfg.System.Http.Timeout),
	}
}
