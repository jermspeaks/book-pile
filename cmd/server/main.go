package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jermspeaks/book-pile/internal/api"
	"github.com/jermspeaks/book-pile/internal/store"
)

func main() {
	addr := getenv("BOOKPILE_ADDR", "127.0.0.1:8080")
	dataDir := getenv("BOOKPILE_DATA", "data")
	webDist := getenv("BOOKPILE_WEB", filepath.Join("web", "dist"))

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	handler := api.New(db, dataDir, webDist)
	log.Printf("book pile listening on http://%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
