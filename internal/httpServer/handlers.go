package httpServer

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
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoClusterNameProvided))
		return
	}
	namespace, ok := vars["namespace"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoNamespaceProvided))
		return
	}
	jobsScans := s.storage.GetJobsScans(clusterName, namespace)
	err := json.NewEncoder(w).Encode(jobsScans)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
}

func (s *HttpServer) getServicesScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoClusterNameProvided))
		return
	}
	namespace, ok := vars["namespace"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoNamespaceProvided))
		return
	}
	servicesScans := s.storage.GetServicesScans(clusterName, namespace)
	err := json.NewEncoder(w).Encode(servicesScans)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
}

func (s *HttpServer) getAllClusters(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	clusters, err := s.storage.GetAllClusters()
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	err = json.NewEncoder(w).Encode(clusters)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
}

func (s *HttpServer) getCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoClusterNameProvided))
		return
	}
	cluster, err := s.storage.GetClusterByName(clusterName)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	err = json.NewEncoder(w).Encode(cluster)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
}

func (s *HttpServer) createCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var cluster model.Cluster
	err := json.NewDecoder(r.Body).Decode(&cluster)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	addedCluster, err := s.storage.AddCluster(&cluster)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	err = json.NewEncoder(w).Encode(addedCluster)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
}

func (s *HttpServer) deleteCluster(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoClusterNameProvided))
		return
	}
	err := s.storage.DeleteCluster(clusterName)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HttpServer) addNamespace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoClusterNameProvided))
		return
	}
	var namespaceStruct namespaceRequestStruct
	err := json.NewDecoder(r.Body).Decode(&namespaceStruct)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	err = s.storage.AddNamespaceToCluster(clusterName, namespaceStruct.Namespace)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HttpServer) deleteNamespace(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoClusterNameProvided))
		return
	}
	namespace, ok := vars["namespace"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoNamespaceProvided))
		return
	}
	err := s.storage.DeleteNamespaceFromCluster(clusterName, namespace)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *HttpServer) changeClusterConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	clusterName, ok := vars["cluster"]
	if !ok {
		s.writeErrorResponse(w, model.NewServerErrorByCode(model.NoClusterNameProvided))
		return
	}
	var clusterConfigStruct clusterConfigRequestStruct
	err := json.NewDecoder(r.Body).Decode(&clusterConfigStruct)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	cluster, err := s.storage.EditClusterConfig(clusterName, clusterConfigStruct.Config)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
	err = json.NewEncoder(w).Encode(cluster)
	if err != nil {
		s.writeErrorResponse(w, err)
		return
	}
}
