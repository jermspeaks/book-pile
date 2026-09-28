package api

import (
	"net/http"
	"strconv"

	"github.com/jermspeaks/book-pile/internal/store"
	"github.com/jermspeaks/book-pile/internal/suggest"
)

func (s *Server) suggestNext(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	unread, err := s.DB.Q.ListUnreadBooks(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	links, err := s.DB.Q.ListBookAuthors(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	authorIDs := map[int64][]int64{}
	for _, l := range links {
		authorIDs[l.BookID] = append(authorIDs[l.BookID], l.AuthorID)
	}
	readRows, err := s.DB.Q.ListReadCountByAuthor(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	readCount := map[int64]int{}
	for _, row := range readRows {
		readCount[row.AuthorID] = int(row.ReadCount)
	}
	ranked := suggest.RankNext(unread, authorIDs, readCount)
	limit := 12
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > len(ranked) {
		limit = len(ranked)
	}
	books := make([]store.Book, 0, limit)
	for i := 0; i < limit; i++ {
		books = append(books, ranked[i].Book)
	}
	hydrated, err := s.hydrateList(r, books)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]SuggestedBook, 0, len(hydrated))
	for i, b := range hydrated {
		out = append(out, SuggestedBook{BookResponse: b, Score: ranked[i].Score, Reason: ranked[i].Reason})
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) suggestAuthors(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Q.ListAuthors(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	limit := 12
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	out := make([]AuthorResponse, 0, limit)
	for _, row := range rows {
		a := toAuthor(row)
		if a.UnreadWantSum == 0 && a.ReadCount == 0 {
			continue
		}
		out = append(out, a)
		if len(out) >= limit {
			break
		}
	}
	writeJSON(w, http.StatusOK, out)
}
