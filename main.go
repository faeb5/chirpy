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
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Print("Unable to load DB_URL")
		return
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Failed to connect to database: %s", err)
	}
	dbQueries := database.New(db)

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		fmt.Println("Unable to load PLATFORM")
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("Unable to load JWT_SECRET")
		return
	}

	const port string = "8080"
	const root string = "."

	apiConfig := apiConfig{
		fileServerHits: atomic.Int32{},
		dbQueries:      dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", apiConfig.middleWareMetricsInc(http.FileServer(http.Dir(root)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetricsShow)
	mux.HandleFunc("POST /admin/reset", apiConfig.handlerReset)
	mux.HandleFunc("POST /api/users", apiConfig.handlerCreateUser)
	mux.HandleFunc("PUT /api/users", apiConfig.handlerUpdateUser)
	mux.HandleFunc("POST /api/chirps", apiConfig.handlerCreateChirp)
	mux.HandleFunc("GET /api/chirps", apiConfig.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handlerGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConfig.handlerDeleteChirp)
	mux.HandleFunc("POST /api/login", apiConfig.handlerLogin)
	mux.HandleFunc("POST /api/revoke", apiConfig.handlerRevokeRefreshToken)
	mux.HandleFunc("POST /api/refresh", apiConfig.handlerUpdateRefreshToken)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	fmt.Printf("Serving files from %s on port %s\n", root, port)
	log.Fatal(server.ListenAndServe())
}
