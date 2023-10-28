package http

import (
	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"scan_project/internal/model"
)

func (s *HttpServer) getJobsScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.handleError(w, model.ServerError{
			Code:        model.WrongRequestFormat,
			Description: "no cluster name provided in request",
		})
		return
	}
	namespace, ok := vars["namespace"]
	if !ok {
		s.handleError(w, model.ServerError{
			Code:        model.WrongRequestFormat,
			Description: "no namespace provided in request",
		})
		return
	}
	jobsScans := s.storage.GetJobsScans(clusterName, namespace)
	err := json.NewEncoder(w).Encode(jobsScans)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *HttpServer) getServicesScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.handleError(w, model.ServerError{
			Code:        model.WrongRequestFormat,
			Description: "no cluster name provided in request",
		})
		return
	}
	namespace, ok := vars["namespace"]
	if !ok {
		s.handleError(w, model.ServerError{
			Code:        model.WrongRequestFormat,
			Description: "no namespace provided in request",
		})
		return
	}
	servicesScans := s.storage.GetServicesScans(clusterName, namespace)
	err := json.NewEncoder(w).Encode(servicesScans)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
