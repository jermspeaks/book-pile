package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jermspeaks/book-pile/internal/openlibrary"
	"github.com/jermspeaks/book-pile/internal/store"
)

type Server struct {
	DB       *store.DB
	DataDir  string
	OL       *openlibrary.Client
	WebDist  string
}

func New(db *store.DB, dataDir, webDist string) http.Handler {
	s := &Server{
		DB:      db,
		DataDir: dataDir,
		OL:      openlibrary.New(),
		WebDist: webDist,
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)
	r.Use(middleware.RealIP)

	r.Handle("/covers/*", http.StripPrefix("/covers/", http.FileServer(http.Dir(filepath.Join(dataDir, "covers")))))

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.SetHeader("Content-Type", "application/json"))
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		})
		r.Get("/books", s.listBooks)
		r.Post("/books", s.createBook)
		r.Get("/books/{id}", s.getBook)
		r.Patch("/books/{id}", s.patchBook)
		r.Put("/books/{id}/piles", s.putBookPiles)
		r.Put("/books/{id}/collections", s.putBookCollections)
		r.Put("/books/{id}/genres", s.putBookGenres)

		r.Get("/piles", s.listPiles)
		r.Post("/piles", s.createPile)
		r.Get("/piles/{id}", s.getPile)

		r.Get("/collections", s.listCollections)
		r.Post("/collections", s.createCollection)
		r.Get("/collections/{id}", s.getCollection)

		r.Get("/genres", s.listGenres)
		r.Get("/genres/{id}", s.getGenre)

		r.Get("/authors", s.listAuthors)
		r.Get("/authors/{id}", s.getAuthor)

		r.Post("/import/calibre/preview", s.previewCalibre)
		r.With(middleware.Timeout(30 * time.Minute)).Post("/import/calibre", s.importCalibre)

		r.Get("/suggest/next", s.suggestNext)
		r.Get("/suggest/authors", s.suggestAuthors)
	})

	if webDist != "" {
		if _, err := os.Stat(webDist); err == nil {
			r.Handle("/*", spa(webDist))
		}
	}
	return r
}

func spa(dist string) http.Handler {
	fileServer := http.FileServer(http.Dir(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		p := filepath.Join(dist, rel)
		if rel != "" && !strings.HasPrefix(p, dist) {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(p)
		if err == nil && !info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dist, "index.html"))
	})
}

func coversDir(dataDir string) string {
	return filepath.Join(dataDir, "covers")
}
