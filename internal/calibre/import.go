package calibre

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jermspeaks/book-pile/internal/store"
	_ "modernc.org/sqlite"
)

type Result struct {
	Imported int `json:"imported"`
	Updated  int `json:"updated"`
	Covers   int `json:"covers"`
	Books    int `json:"books"`
}

type calibreBook struct {
	ID          int64
	Title       string
	SeriesIndex float64
	Path        string
	HasCover    bool
	Pubdate     string
	ISBN        string
	Authors     []string
	Tags        []string
	Comment     string
	Series      string
	Identifiers map[string]string
	Rating      *float64
	Formats     []string
}

func Preview(libraryPath string) (int, error) {
	db, err := openCalibre(libraryPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM books`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count calibre books: %w", err)
	}
	return n, nil
}

func Import(ctx context.Context, db *store.DB, libraryPath, coversDir string) (Result, error) {
	src, err := openCalibre(libraryPath)
	if err != nil {
		return Result{}, err
	}
	defer src.Close()

	books, err := loadCalibre(src)
	if err != nil {
		return Result{}, err
	}

	tx, err := db.SQL.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer func() { _ = tx.Rollback() }()
	q := db.Q.WithTx(tx)

	calibrePile, err := store.EnsurePile(ctx, q, "Calibre")
	if err != nil {
		return Result{}, err
	}

	var res Result
	res.Books = len(books)
	for _, cb := range books {
		if err := ctx.Err(); err != nil {
			return Result{}, err
		}
		book, created, err := upsertBook(ctx, q, cb)
		if err != nil {
			return Result{}, fmt.Errorf("upsert %q: %w", cb.Title, err)
		}
		if created {
			res.Imported++
		} else {
			res.Updated++
		}
		if err := replaceAuthors(ctx, q, book.ID, cb.Authors); err != nil {
			return Result{}, err
		}
		if err := addGenres(ctx, q, book.ID, cb.Tags); err != nil {
			return Result{}, err
		}
		if err := q.InsertBookPile(ctx, store.InsertBookPileParams{BookID: book.ID, PileID: calibrePile.ID}); err != nil {
			return Result{}, err
		}
		if cb.HasCover {
			srcCover := filepath.Join(libraryPath, cb.Path, "cover.jpg")
			if copied, err := copyCover(srcCover, coversDir, book.ID); err != nil {
				return Result{}, err
			} else if copied {
				name := fmt.Sprintf("%d.jpg", book.ID)
				if err := q.SetCoverPath(ctx, store.SetCoverPathParams{CoverPath: &name, ID: book.ID}); err != nil {
					return Result{}, err
				}
				res.Covers++
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return Result{}, err
	}
	return res, nil
}

func openCalibre(libraryPath string) (*sql.DB, error) {
	meta := filepath.Join(libraryPath, "metadata.db")
	if _, err := os.Stat(meta); err != nil {
		return nil, fmt.Errorf("calibre library not found: expected metadata.db in %s", libraryPath)
	}
	dsn := fmt.Sprintf("file:%s?mode=ro&_pragma=query_only(1)", filepath.ToSlash(meta))
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open calibre db: %w", err)
	}
	return db, nil
}

func loadCalibre(db *sql.DB) ([]calibreBook, error) {
	rows, err := db.Query(`SELECT id, title, series_index, path, has_cover, COALESCE(pubdate, ''), COALESCE(isbn, '') FROM books`)
	if err != nil {
		return nil, fmt.Errorf("read books: %w", err)
	}
	defer rows.Close()
	byID := map[int64]*calibreBook{}
	var ordered []int64
	for rows.Next() {
		var b calibreBook
		var hasCover int64
		if err := rows.Scan(&b.ID, &b.Title, &b.SeriesIndex, &b.Path, &hasCover, &b.Pubdate, &b.ISBN); err != nil {
			return nil, err
		}
		b.HasCover = hasCover != 0
		b.Identifiers = map[string]string{}
		byID[b.ID] = &b
		ordered = append(ordered, b.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	authors, err := mapNames(db, `SELECT id, name FROM authors`)
	if err != nil {
		return nil, err
	}
	if err := linkMany(db, `SELECT book, author FROM books_authors_link`, byID, func(b *calibreBook, id int64) {
		if n, ok := authors[id]; ok {
			b.Authors = append(b.Authors, n)
		}
	}); err != nil {
		return nil, err
	}

	tags, err := mapNames(db, `SELECT id, name FROM tags`)
	if err != nil && !isMissing(err) {
		return nil, err
	}
	if err == nil {
		_ = linkMany(db, `SELECT book, tag FROM books_tags_link`, byID, func(b *calibreBook, id int64) {
			if n, ok := tags[id]; ok {
				b.Tags = append(b.Tags, n)
			}
		})
	}

	series, err := mapNames(db, `SELECT id, name FROM series`)
	if err != nil && !isMissing(err) {
		return nil, err
	}
	if err == nil {
		_ = linkMany(db, `SELECT book, series FROM books_series_link`, byID, func(b *calibreBook, id int64) {
			if n, ok := series[id]; ok {
				b.Series = n
			}
		})
	}

	if cRows, err := db.Query(`SELECT book, text FROM comments`); err == nil {
		for cRows.Next() {
			var bookID int64
			var text string
			if err := cRows.Scan(&bookID, &text); err != nil {
				cRows.Close()
				return nil, err
			}
			if b := byID[bookID]; b != nil {
				b.Comment = text
			}
		}
		cRows.Close()
	}

	if iRows, err := db.Query(`SELECT book, type, val FROM identifiers`); err == nil {
		for iRows.Next() {
			var bookID int64
			var typ, val string
			if err := iRows.Scan(&bookID, &typ, &val); err != nil {
				iRows.Close()
				return nil, err
			}
			if b := byID[bookID]; b != nil {
				b.Identifiers[strings.ToLower(typ)] = val
			}
		}
		iRows.Close()
	}

	ratings := map[int64]float64{}
	if rRows, err := db.Query(`SELECT id, rating FROM ratings`); err == nil {
		for rRows.Next() {
			var id int64
			var rating float64
			if err := rRows.Scan(&id, &rating); err != nil {
				rRows.Close()
				return nil, err
			}
			ratings[id] = rating
		}
		rRows.Close()
		_ = linkMany(db, `SELECT book, rating FROM books_ratings_link`, byID, func(b *calibreBook, id int64) {
			if r, ok := ratings[id]; ok {
				v := r
				b.Rating = &v
			}
		})
	}

	if dRows, err := db.Query(`SELECT book, format FROM data`); err == nil {
		for dRows.Next() {
			var bookID int64
			var format string
			if err := dRows.Scan(&bookID, &format); err != nil {
				dRows.Close()
				return nil, err
			}
			if b := byID[bookID]; b != nil {
				b.Formats = append(b.Formats, format)
			}
		}
		dRows.Close()
	}

	out := make([]calibreBook, 0, len(ordered))
	for _, id := range ordered {
		out = append(out, *byID[id])
	}
	return out, nil
}

func upsertBook(ctx context.Context, q *store.Queries, cb calibreBook) (store.Book, bool, error) {
	isbn := cb.ISBN
	if v, ok := cb.Identifiers["isbn"]; ok && v != "" {
		isbn = v
	}
	isbn = strings.ReplaceAll(strings.TrimSpace(isbn), "-", "")
	year := parseYear(cb.Pubdate)
	calibreID := cb.ID
	params := store.InsertBookParams{
		Title:         strings.TrimSpace(cb.Title),
		Isbn:          store.StrPtr(isbn),
		PublishedYear: year,
		Description:   store.StrPtr(stripHTML(cb.Comment)),
		SeriesName:    store.StrPtr(cb.Series),
		SeriesIndex:   floatPtr(cb.SeriesIndex),
		OwnedPhysical: 0,
		OwnedDigital:  1,
		ReadStatus:    "unread",
		QualityRating: cb.Rating,
		CalibreID:     &calibreID,
	}

	existing, err := q.GetBookByCalibreID(ctx, &calibreID)
	if err == sql.ErrNoRows {
		existing, err = matchExisting(ctx, q, isbn, cb.Title, cb.Authors)
	}
	if err == nil {
		updated, uerr := q.UpdateCalibreMetadata(ctx, store.UpdateCalibreMetadataParams{
			Title:         params.Title,
			Isbn:          params.Isbn,
			PublishedYear: params.PublishedYear,
			Description:   params.Description,
			SeriesName:    params.SeriesName,
			SeriesIndex:   params.SeriesIndex,
			QualityRating: params.QualityRating,
			CalibreID:     params.CalibreID,
			ID:            existing.ID,
		})
		return updated, false, uerr
	}
	if err != sql.ErrNoRows {
		return store.Book{}, false, err
	}
	created, err := q.InsertBook(ctx, params)
	return created, true, err
}

func matchExisting(ctx context.Context, q *store.Queries, isbn, title string, authors []string) (store.Book, error) {
	if isbn != "" {
		b, err := q.GetBookByISBN(ctx, store.StrPtr(isbn))
		if err == nil {
			return b, nil
		}
		if err != sql.ErrNoRows {
			return store.Book{}, err
		}
	}
	candidates, err := q.ListBooksByNormalizedTitle(ctx, title)
	if err != nil {
		return store.Book{}, err
	}
	wantAuthors := map[string]struct{}{}
	for _, a := range authors {
		wantAuthors[store.Normalize(a)] = struct{}{}
	}
	for _, b := range candidates {
		links, err := q.ListAuthorsForBooks(ctx, []int64{b.ID})
		if err != nil {
			return store.Book{}, err
		}
		for _, l := range links {
			if _, ok := wantAuthors[store.Normalize(l.Name)]; ok {
				return b, nil
			}
		}
	}
	return store.Book{}, sql.ErrNoRows
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

func addGenres(ctx context.Context, q *store.Queries, bookID int64, tags []string) error {
	for _, name := range tags {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		g, err := store.EnsureGenre(ctx, q, name)
		if err != nil {
			return err
		}
		if err := q.InsertBookGenre(ctx, store.InsertBookGenreParams{BookID: bookID, GenreID: g.ID}); err != nil {
			return err
		}
	}
	return nil
}

func copyCover(src, coversDir string, bookID int64) (bool, error) {
	in, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	defer in.Close()
	if err := os.MkdirAll(coversDir, 0o755); err != nil {
		return false, err
	}
	dest := filepath.Join(coversDir, fmt.Sprintf("%d.jpg", bookID))
	out, err := os.Create(dest)
	if err != nil {
		return false, err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return false, err
	}
	return true, nil
}

func mapNames(db *sql.DB, q string) (map[int64]string, error) {
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64]string{}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		out[id] = name
	}
	return out, rows.Err()
}

func linkMany(db *sql.DB, q string, books map[int64]*calibreBook, add func(*calibreBook, int64)) error {
	rows, err := db.Query(q)
	if err != nil {
		if isMissing(err) {
			return nil
		}
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var bookID, otherID int64
		if err := rows.Scan(&bookID, &otherID); err != nil {
			return err
		}
		if b := books[bookID]; b != nil {
			add(b, otherID)
		}
	}
	return rows.Err()
}

func isMissing(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "no such table")
}

func parseYear(pubdate string) *int64 {
	pubdate = strings.TrimSpace(pubdate)
	if pubdate == "" {
		return nil
	}
	if t, err := time.Parse("2006-01-02", pubdate[:min(10, len(pubdate))]); err == nil {
		y := int64(t.Year())
		if y > 1 {
			return &y
		}
	}
	if y, err := strconv.ParseInt(pubdate[:min(4, len(pubdate))], 10, 64); err == nil && y > 1 {
		return &y
	}
	return nil
}

func floatPtr(v float64) *float64 { return &v }

func stripHTML(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	return strings.TrimSpace(b.String())
}
