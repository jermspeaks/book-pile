package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/jermspeaks/book-pile/internal/store"
)

type Named struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type BookResponse struct {
	ID             int64    `json:"id"`
	Title          string   `json:"title"`
	Subtitle       *string  `json:"subtitle"`
	ISBN           *string  `json:"isbn"`
	PublishedYear  *int64   `json:"publishedYear"`
	Description    *string  `json:"description"`
	CoverURL       *string  `json:"coverUrl"`
	SeriesName     *string  `json:"seriesName"`
	SeriesIndex    *float64 `json:"seriesIndex"`
	OwnedPhysical  bool     `json:"ownedPhysical"`
	OwnedDigital   bool     `json:"ownedDigital"`
	Ownership      string   `json:"ownership"`
	ReadStatus     string   `json:"readStatus"`
	WantRating     *int64   `json:"wantRating"`
	QualityRating  *float64 `json:"qualityRating"`
	Authors        []Named  `json:"authors"`
	Piles          []Named  `json:"piles,omitempty"`
	Collections    []Named  `json:"collections,omitempty"`
	Genres         []Named  `json:"genres,omitempty"`
	FirstAuthor    string   `json:"-"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

type SuggestedBook struct {
	BookResponse
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type AuthorResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	BookCount      int64  `json:"bookCount"`
	ReadCount      int64  `json:"readCount"`
	UnreadWantSum  int64  `json:"unreadWantSum"`
	UnreadOwned    int64  `json:"unreadOwned"`
	UnreadCoveting int64  `json:"unreadCoveting"`
}

type AuthorDetail struct {
	AuthorResponse
	Owned    []BookResponse `json:"owned"`
	Coveting []BookResponse `json:"coveting"`
}

func toBookResponse(b store.Book, authors []Named) BookResponse {
	return BookResponse{
		ID:            b.ID,
		Title:         b.Title,
		Subtitle:      b.Subtitle,
		ISBN:          b.Isbn,
		PublishedYear: b.PublishedYear,
		Description:   b.Description,
		CoverURL:      store.CoverURL(b.CoverPath),
		SeriesName:    b.SeriesName,
		SeriesIndex:   b.SeriesIndex,
		OwnedPhysical: b.OwnedPhysical != 0,
		OwnedDigital:  b.OwnedDigital != 0,
		Ownership:     store.Ownership(b.OwnedPhysical, b.OwnedDigital),
		ReadStatus:    b.ReadStatus,
		WantRating:    b.WantRating,
		QualityRating: b.QualityRating,
		Authors:       nonNil(authors),
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func queryInt64(r *http.Request, key string) any {
	s := r.URL.Query().Get(key)
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return n
}

func queryString(r *http.Request, key string) any {
	s := r.URL.Query().Get(key)
	if s == "" {
		return nil
	}
	return s
}

func named(id int64, name string) Named { return Named{ID: id, Name: name} }
