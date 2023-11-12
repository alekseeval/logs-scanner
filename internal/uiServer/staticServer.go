package staticServer

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"scan_project/configuration"
	"time"
)

type fileServer struct {
	logger *logrus.Entry
}

func NewFileServer(cfg configuration.Config, logger *logrus.Entry) *http.Server {
	staticServer := fileServer{
		logger: logger,
	}
	r := mux.NewRouter()
	r.Use(staticServer.loggingMiddleware)
	// Swagger UI
	sh := http.StripPrefix("/swagger/", http.FileServer(http.Dir("./static/swaggerui/")))
	r.PathPrefix("/swagger/").Handler(sh)
	return &http.Server{
		Addr:         fmt.Sprintf(":%d", 8080), // TODO: Задавать в конфиге
		Handler:      r,
		ReadTimeout:  time.Second * time.Duration(cfg.System.Http.Timeout),
		WriteTimeout: time.Second * time.Duration(cfg.System.Http.Timeout),
	}
}
