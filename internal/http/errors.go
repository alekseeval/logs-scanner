package http

import (
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"scan_project/internal/model"
)

func (s *HttpServer) handleError(w http.ResponseWriter, serverError model.ServerError) {
	w.WriteHeader(http.StatusBadRequest)
	err := json.NewEncoder(w).Encode(serverError)
	if err != nil {
		s.logger.
			WithField("error", err).
			Error("Failed to marshall error response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
