package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func main() {
	const port string = "8080"
	const root string = "."

	apiConfig := apiConfig{
		fileServerHits: atomic.Int32{},
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiConfig.middleWareMetricsInc(http.FileServer(http.Dir(root)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetricsShow)
	mux.HandleFunc("POST /admin/reset", apiConfig.handlerMetricsReset)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Serving files from %s on port %s\n", root, port)
	log.Fatal(server.ListenAndServe())
}
