package httpServer

import (
	"k8s.io/apimachinery/pkg/util/json"
	"net/http"
	"scan_project/internal/model"
)

func (s *HttpServer) writeErrorResponse(w http.ResponseWriter, externalErr error) {
	w.WriteHeader(http.StatusBadRequest)
	var err error
	switch externalErr.(type) {
	case *model.ServerError:
		err = json.NewEncoder(w).Encode(externalErr)
	default:
		err = json.NewEncoder(w).Encode(model.ServerError{
			Code:        model.InternalServerError,
			Description: err.Error(),
		})
	}
	if err != nil {
		s.logger.
			WithField("error", err).
			Error("Failed to marshall error response")
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
