package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jermspeaks/book-pile/internal/store"
)

type nameBody struct {
	Name string `json:"name"`
}

func (s *Server) listPiles(w http.ResponseWriter, r *http.Request) {
	items, err := s.DB.Q.ListPiles(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nonNil(items))
}

func (s *Server) getPile(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := s.DB.Q.GetPile(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "pile not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) createPile(w http.ResponseWriter, r *http.Request) {
	var body nameBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	item, err := store.EnsurePile(r.Context(), s.DB.Q, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) listCollections(w http.ResponseWriter, r *http.Request) {
	items, err := s.DB.Q.ListCollections(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nonNil(items))
}

func (s *Server) getCollection(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := s.DB.Q.GetCollection(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "collection not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) createCollection(w http.ResponseWriter, r *http.Request) {
	var body nameBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	item, err := store.EnsureCollection(r.Context(), s.DB.Q, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) listGenres(w http.ResponseWriter, r *http.Request) {
	items, err := s.DB.Q.ListGenres(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nonNil(items))
}

func (s *Server) getGenre(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	item, err := s.DB.Q.GetGenre(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "genre not found")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *Server) listAuthors(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Q.ListAuthors(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]AuthorResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAuthor(row))
	}
	writeJSON(w, http.StatusOK, nonNil(out))
}

func (s *Server) getAuthor(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctx := r.Context()
	author, err := s.DB.Q.GetAuthor(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "author not found")
		return
	}
	rows, err := s.DB.Q.ListAuthors(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	detail := AuthorDetail{AuthorResponse: AuthorResponse{ID: author.ID, Name: author.Name}}
	for _, row := range rows {
		if row.ID == author.ID {
			detail.AuthorResponse = toAuthor(row)
			break
		}
	}
	books, err := s.DB.Q.ListBooksByAuthor(ctx, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	hydrated, err := s.hydrateList(r, books)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	detail.Owned = []BookResponse{}
	detail.Coveting = []BookResponse{}
	for _, b := range hydrated {
		if b.Ownership == "coveting" {
			detail.Coveting = append(detail.Coveting, b)
		} else {
			detail.Owned = append(detail.Owned, b)
		}
	}
	writeJSON(w, http.StatusOK, detail)
}

func toAuthor(row store.ListAuthorsRow) AuthorResponse {
	return AuthorResponse{
		ID:             row.ID,
		Name:           row.Name,
		BookCount:      row.BookCount,
		ReadCount:      row.ReadCount,
		UnreadWantSum:  row.UnreadWantSum,
		UnreadOwned:    row.UnreadOwned,
		UnreadCoveting: row.UnreadCoveting,
	}
}
