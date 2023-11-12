package staticServer

import "net/http"

func (s *fileServer) loggingMiddleware(next http.Handler) http.Handler {
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
