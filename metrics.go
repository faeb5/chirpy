package main

import (
	"fmt"
	"net/http"
)

func (config *apiConfig) middleWareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		config.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (config *apiConfig) handlerMetricsShow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, config.fileServerHits.Load())))
}

func (config *apiConfig) handlerMetricsReset(w http.ResponseWriter, r *http.Request) {
	config.fileServerHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits resetted to 0"))
}
