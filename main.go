package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"example.com/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileServerHits atomic.Int32
    dbQueries *database.Queries
}

func main() {
	godotenv.Load()
    dbURL := os.Getenv("DB_URL")
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        fmt.Errorf("Failed to connect to database: %s", err)
    }

	const port string = "8080"
	const root string = "."

	apiConfig := apiConfig{
		fileServerHits: atomic.Int32{},
        dbQueries: database.New(db),
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
