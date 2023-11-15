package httpServer

import (
	"net/http"
)

func (s *httpServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.
			WithField("uri", r.URL.Path).
			WithField("headers", r.Header).
			WithField("body", r.Body).
			WithField("method", r.Method).
			Info("Request handled")
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func setResponseHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PATCH, DELETE, PUT")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Access-Control-Allow-Origin, "+
			"Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, "+
			"Access-Control-Request-Method, Access-Control-Request-Headers,X-Auth-Token,"+
			"X-Additional-Encryption-Enabled,X-cd-val")
		if r.Method == http.MethodOptions {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
