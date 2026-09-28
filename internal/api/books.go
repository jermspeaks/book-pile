package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jermspeaks/book-pile/internal/store"
)

func (s *Server) listBooks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()
	books, err := s.DB.Q.ListBooks(ctx, store.ListBooksParams{
		ReadStatus:   queryString(r, "readStatus"),
		MinWant:      queryInt64(r, "minWant"),
		PileID:       queryInt64(r, "pileId"),
		CollectionID: queryInt64(r, "collectionId"),
		GenreID:      queryInt64(r, "genreId"),
		AuthorID:     queryInt64(r, "authorId"),
		Ownership:    queryString(r, "ownership"),
		Q:            queryString(r, "q"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out, err := s.hydrateList(r, books)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sortBooks(out, q.Get("sort"))
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) getBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	b, err := s.DB.Q.GetBook(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	out, err := s.hydrateOne(r, b)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type createBookRequest struct {
	Title         string   `json:"title"`
	Subtitle      *string  `json:"subtitle"`
	Authors       []string `json:"authors"`
	ISBN          *string  `json:"isbn"`
	PublishedYear *int64   `json:"publishedYear"`
	Description   *string  `json:"description"`
	OwnedPhysical bool     `json:"ownedPhysical"`
	OwnedDigital  bool     `json:"ownedDigital"`
	ReadStatus    string   `json:"readStatus"`
	WantRating    *int64   `json:"wantRating"`
	FetchMetadata bool     `json:"fetchMetadata"`
}

func (s *Server) createBook(w http.ResponseWriter, r *http.Request) {
	var req createBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	title := strings.TrimSpace(req.Title)
	isbn := ""
	if req.ISBN != nil {
		isbn = strings.ReplaceAll(strings.TrimSpace(*req.ISBN), "-", "")
	}
	if req.FetchMetadata && isbn != "" {
		if meta, err := s.OL.LookupISBN(isbn); err == nil {
			if title == "" {
				title = meta.Title
			}
			if len(req.Authors) == 0 {
				req.Authors = meta.Authors
			}
			if req.PublishedYear == nil {
				req.PublishedYear = meta.Year
			}
		}
	}
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	readStatus := req.ReadStatus
	if readStatus == "" {
		readStatus = "unread"
	}
	ctx := r.Context()
	book, err := s.DB.Q.InsertBook(ctx, store.InsertBookParams{
		Title:         title,
		Subtitle:      req.Subtitle,
		Isbn:          store.StrPtr(isbn),
		PublishedYear: req.PublishedYear,
		Description:   req.Description,
		OwnedPhysical: store.BoolToInt(req.OwnedPhysical),
		OwnedDigital:  store.BoolToInt(req.OwnedDigital),
		ReadStatus:    readStatus,
		WantRating:    req.WantRating,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if err := replaceAuthors(ctx, s.DB.Q, book.ID, req.Authors); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if isbn != "" {
		dest := fmt.Sprintf("%s/%d.jpg", coversDir(s.DataDir), book.ID)
		if ok, err := s.OL.DownloadCover(isbn, dest); err == nil && ok {
			name := fmt.Sprintf("%d.jpg", book.ID)
			_ = s.DB.Q.SetCoverPath(ctx, store.SetCoverPathParams{CoverPath: &name, ID: book.ID})
			book.CoverPath = &name
		}
	}
	out, err := s.hydrateOne(r, book)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, out)
}

type patchBookRequest struct {
	Title         *string `json:"title"`
	Subtitle      *string `json:"subtitle"`
	ISBN          *string `json:"isbn"`
	PublishedYear *int64  `json:"publishedYear"`
	Description   *string `json:"description"`
	SeriesName    *string `json:"seriesName"`
	WantRating    *int64  `json:"wantRating"`
	ReadStatus    *string `json:"readStatus"`
	OwnedPhysical *bool   `json:"ownedPhysical"`
	OwnedDigital  *bool   `json:"ownedDigital"`
	Authors       []string `json:"authors"`
}

func (s *Server) patchBook(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ctx := r.Context()
	book, err := s.DB.Q.GetBook(ctx, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "book not found")
		return
	}
	var req patchBookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Title != nil {
		book.Title = strings.TrimSpace(*req.Title)
	}
	if req.Subtitle != nil {
		book.Subtitle = req.Subtitle
	}
	if req.ISBN != nil {
		v := strings.ReplaceAll(strings.TrimSpace(*req.ISBN), "-", "")
		book.Isbn = store.StrPtr(v)
	}
	if req.PublishedYear != nil {
		book.PublishedYear = req.PublishedYear
	}
	if req.Description != nil {
		book.Description = req.Description
	}
	if req.SeriesName != nil {
		book.SeriesName = req.SeriesName
	}
	if req.WantRating != nil {
		book.WantRating = req.WantRating
	}
	if req.ReadStatus != nil {
		book.ReadStatus = *req.ReadStatus
	}
	if req.OwnedPhysical != nil {
		book.OwnedPhysical = store.BoolToInt(*req.OwnedPhysical)
	}
	if req.OwnedDigital != nil {
		book.OwnedDigital = store.BoolToInt(*req.OwnedDigital)
	}
	updated, err := s.DB.Q.UpdateBookUserFields(ctx, store.UpdateBookUserFieldsParams{
		Title:         book.Title,
		Subtitle:      book.Subtitle,
		Isbn:          book.Isbn,
		PublishedYear: book.PublishedYear,
		Description:   book.Description,
		SeriesName:    book.SeriesName,
		WantRating:    book.WantRating,
		ReadStatus:    book.ReadStatus,
		OwnedPhysical: book.OwnedPhysical,
		OwnedDigital:  book.OwnedDigital,
		ID:            book.ID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if req.Authors != nil {
		if err := replaceAuthors(ctx, s.DB.Q, updated.ID, req.Authors); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	out, err := s.hydrateOne(r, updated)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, out)
}

type idsBody struct {
	PileIDs       []int64 `json:"pileIds"`
	CollectionIDs []int64 `json:"collectionIds"`
	GenreIDs      []int64 `json:"genreIds"`
}

func (s *Server) putBookPiles(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body idsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx := r.Context()
	if err := s.DB.Q.DeleteBookPiles(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, pileID := range body.PileIDs {
		if err := s.DB.Q.InsertBookPile(ctx, store.InsertBookPileParams{BookID: id, PileID: pileID}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	s.getBook(w, r)
}

func (s *Server) putBookCollections(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body idsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx := r.Context()
	if err := s.DB.Q.DeleteBookCollections(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, cid := range body.CollectionIDs {
		if err := s.DB.Q.InsertBookCollection(ctx, store.InsertBookCollectionParams{BookID: id, CollectionID: cid}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	s.getBook(w, r)
}

func (s *Server) putBookGenres(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body idsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	ctx := r.Context()
	if err := s.DB.Q.DeleteBookGenres(ctx, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, gid := range body.GenreIDs {
		if err := s.DB.Q.InsertBookGenre(ctx, store.InsertBookGenreParams{BookID: id, GenreID: gid}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	s.getBook(w, r)
}

func (s *Server) hydrateList(r *http.Request, books []store.Book) ([]BookResponse, error) {
	if len(books) == 0 {
		return []BookResponse{}, nil
	}
	ids := make([]int64, len(books))
	for i, b := range books {
		ids[i] = b.ID
	}
	links, err := s.DB.Q.ListAuthorsForBooks(r.Context(), ids)
	if err != nil {
		return nil, err
	}
	byBook := map[int64][]Named{}
	for _, l := range links {
		byBook[l.BookID] = append(byBook[l.BookID], named(l.AuthorID, l.Name))
	}
	out := make([]BookResponse, 0, len(books))
	for _, b := range books {
		authors := byBook[b.ID]
		if authors == nil {
			authors = []Named{}
		}
		resp := toBookResponse(b, authors)
		if len(authors) > 0 {
			resp.FirstAuthor = authors[0].Name
		}
		out = append(out, resp)
	}
	return out, nil
}

func (s *Server) hydrateOne(r *http.Request, b store.Book) (BookResponse, error) {
	ctx := r.Context()
	links, err := s.DB.Q.ListAuthorsForBooks(ctx, []int64{b.ID})
	if err != nil {
		return BookResponse{}, err
	}
	namedAuthors := make([]Named, 0, len(links))
	for _, l := range links {
		namedAuthors = append(namedAuthors, named(l.AuthorID, l.Name))
	}
	resp := toBookResponse(b, namedAuthors)
	piles, err := s.DB.Q.ListPilesForBook(ctx, b.ID)
	if err != nil {
		return BookResponse{}, err
	}
	resp.Piles = make([]Named, 0, len(piles))
	for _, p := range piles {
		resp.Piles = append(resp.Piles, named(p.ID, p.Name))
	}
	cols, err := s.DB.Q.ListCollectionsForBook(ctx, b.ID)
	if err != nil {
		return BookResponse{}, err
	}
	resp.Collections = make([]Named, 0, len(cols))
	for _, c := range cols {
		resp.Collections = append(resp.Collections, named(c.ID, c.Name))
	}
	genres, err := s.DB.Q.ListGenresForBook(ctx, b.ID)
	if err != nil {
		return BookResponse{}, err
	}
	resp.Genres = make([]Named, 0, len(genres))
	for _, g := range genres {
		resp.Genres = append(resp.Genres, named(g.ID, g.Name))
	}
	if len(namedAuthors) > 0 {
		resp.FirstAuthor = namedAuthors[0].Name
	}
	return resp, nil
}

func replaceAuthors(ctx context.Context, q *store.Queries, bookID int64, names []string) error {
	if err := q.DeleteBookAuthors(ctx, bookID); err != nil {
		return err
	}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		a, err := store.EnsureAuthor(ctx, q, name)
		if err != nil {
			return err
		}
		if err := q.InsertBookAuthor(ctx, store.InsertBookAuthorParams{BookID: bookID, AuthorID: a.ID}); err != nil {
			return err
		}
	}
	return nil
}

func sortBooks(books []BookResponse, sortKey string) {
	sort.SliceStable(books, func(i, j int) bool {
		a, b := books[i], books[j]
		switch sortKey {
		case "title":
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		case "author":
			if a.FirstAuthor != b.FirstAuthor {
				return strings.ToLower(a.FirstAuthor) < strings.ToLower(b.FirstAuthor)
			}
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		case "added":
			return a.CreatedAt > b.CreatedAt
		default:
			aw, bw := int64(-1), int64(-1)
			if a.WantRating != nil {
				aw = *a.WantRating
			}
			if b.WantRating != nil {
				bw = *b.WantRating
			}
			if aw != bw {
				return aw > bw
			}
			return strings.ToLower(a.Title) < strings.ToLower(b.Title)
		}
	})
}

func pathID(r *http.Request) (int64, error) {
	return strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
}
