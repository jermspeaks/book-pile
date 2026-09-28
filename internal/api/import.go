package api

import (
	"encoding/json"
	"net/http"

	"github.com/jermspeaks/book-pile/internal/calibre"
)

type calibrePath struct {
	LibraryPath string `json:"libraryPath"`
}

func (s *Server) previewCalibre(w http.ResponseWriter, r *http.Request) {
	var body calibrePath
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.LibraryPath == "" {
		writeError(w, http.StatusBadRequest, "libraryPath is required")
		return
	}
	n, err := calibre.Preview(body.LibraryPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"books": n})
}

func (s *Server) importCalibre(w http.ResponseWriter, r *http.Request) {
	var body calibrePath
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.LibraryPath == "" {
		writeError(w, http.StatusBadRequest, "libraryPath is required")
		return
	}
	res, err := calibre.Import(r.Context(), s.DB, body.LibraryPath, coversDir(s.DataDir))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, res)
}
