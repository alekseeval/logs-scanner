package http

import (
	"encoding/json"
	"github.com/gorilla/mux"
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
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
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
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
}

func (s *HttpServer) getAllClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	configs, err := s.storage.GetAllConfigs()
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0,
			Description: "TODO",
		})
	}
	err = json.NewEncoder(w).Encode(configs)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
}

func (s *HttpServer) getCluster(w http.ResponseWriter, r *http.Request) {
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
	kubeConfig, err := s.storage.GetKubeConfigByName(clusterName)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0,
			Description: err.Error(),
		})
		return
	}
	err = json.NewEncoder(w).Encode(kubeConfig)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
}

func (s *HttpServer) createCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var cluster model.KubeConfig
	err := json.NewDecoder(r.Body).Decode(&cluster)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        model.WrongRequestFormat,
			Description: err.Error(),
		})
		return
	}
	addedCluster, err := s.storage.AddKubeConfig(&cluster)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0,
			Description: err.Error(),
		})
		return
	}
	err = json.NewEncoder(w).Encode(addedCluster)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
}

func (s *HttpServer) deleteCluster(w http.ResponseWriter, r *http.Request) {
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
	err := s.storage.DeleteKubeConfig(clusterName)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0,
			Description: err.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HttpServer) addNamespace(w http.ResponseWriter, r *http.Request) {
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
	var namespaceStruct namespaceRequestStruct
	err := json.NewDecoder(r.Body).Decode(&namespaceStruct)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
	err = s.storage.AddNamespaceToCubeConfig(clusterName, namespaceStruct.Namespace)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HttpServer) deleteNamespace(w http.ResponseWriter, r *http.Request) {
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
	err := s.storage.DeleteNamespaceFromKubeconfig(clusterName, namespace)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HttpServer) changeClusterConfig(w http.ResponseWriter, r *http.Request) {
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
	var clusterConfigStruct clusterConfigRequestStruct
	err := json.NewDecoder(r.Body).Decode(&clusterConfigStruct)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
	cluster, err := s.storage.EditKubeConfig(clusterName, clusterConfigStruct.Config)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
	err = json.NewEncoder(w).Encode(cluster)
	if err != nil {
		s.handleError(w, model.ServerError{
			Code:        0, // TODO
			Description: err.Error(),
		})
		return
	}
}
