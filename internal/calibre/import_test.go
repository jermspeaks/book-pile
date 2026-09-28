package calibre

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/jermspeaks/book-pile/internal/store"
	_ "modernc.org/sqlite"
)

func TestImportIdempotentAndPreservesUserFields(t *testing.T) {
	lib := t.TempDir()
	if err := writeFixtureLibrary(lib); err != nil {
		t.Fatal(err)
	}
	dataDir := t.TempDir()
	db, err := store.Open(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	first, err := Import(ctx, db, lib, filepath.Join(dataDir, "covers"))
	if err != nil {
		t.Fatal(err)
	}
	if first.Imported != 2 || first.Updated != 0 {
		t.Fatalf("first import: %+v", first)
	}
	if first.Covers != 1 {
		t.Fatalf("expected 1 cover, got %d", first.Covers)
	}

	books, err := db.Q.ListBooks(ctx, store.ListBooksParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(books) != 2 {
		t.Fatalf("books=%d", len(books))
	}
	var target store.Book
	for _, b := range books {
		if b.Title == "The Dispossessed" {
			target = b
		}
	}
	if target.ID == 0 {
		t.Fatal("missing imported book")
	}
	if store.Ownership(target.OwnedPhysical, target.OwnedDigital) != "digital" {
		t.Fatalf("ownership=%s", store.Ownership(target.OwnedPhysical, target.OwnedDigital))
	}
	want := int64(5)
	_, err = db.Q.UpdateBookUserFields(ctx, store.UpdateBookUserFieldsParams{
		Title:         target.Title,
		Subtitle:      target.Subtitle,
		Isbn:          target.Isbn,
		PublishedYear: target.PublishedYear,
		Description:   target.Description,
		SeriesName:    target.SeriesName,
		WantRating:    &want,
		ReadStatus:    "read",
		OwnedPhysical: 1,
		OwnedDigital:  target.OwnedDigital,
		ID:            target.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	second, err := Import(ctx, db, lib, filepath.Join(dataDir, "covers"))
	if err != nil {
		t.Fatal(err)
	}
	if second.Imported != 0 || second.Updated != 2 {
		t.Fatalf("second import: %+v", second)
	}
	got, err := db.Q.GetBook(ctx, target.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.WantRating == nil || *got.WantRating != 5 {
		t.Fatalf("want clobbered: %+v", got.WantRating)
	}
	if got.ReadStatus != "read" {
		t.Fatalf("read status clobbered: %s", got.ReadStatus)
	}
	if got.OwnedPhysical != 1 {
		t.Fatalf("physical ownership clobbered")
	}
	piles, err := db.Q.ListPilesForBook(ctx, got.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(piles) != 1 || piles[0].Name != "Calibre" {
		t.Fatalf("piles=%+v", piles)
	}
}

func writeFixtureLibrary(dir string) error {
	meta := filepath.Join(dir, "metadata.db")
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(meta))
	if err != nil {
		return err
	}
	defer db.Close()
	stmts := []string{
		`CREATE TABLE books (id INTEGER PRIMARY KEY, title TEXT NOT NULL, series_index REAL NOT NULL DEFAULT 1.0, path TEXT NOT NULL, has_cover BOOL DEFAULT 0, pubdate TIMESTAMP, isbn TEXT DEFAULT '')`,
		`CREATE TABLE authors (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`,
		`CREATE TABLE books_authors_link (book INTEGER, author INTEGER)`,
		`CREATE TABLE tags (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`,
		`CREATE TABLE books_tags_link (book INTEGER, tag INTEGER)`,
		`CREATE TABLE comments (id INTEGER PRIMARY KEY, book INTEGER, text TEXT)`,
		`CREATE TABLE identifiers (id INTEGER PRIMARY KEY, book INTEGER, type TEXT, val TEXT)`,
		`CREATE TABLE series (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`,
		`CREATE TABLE books_series_link (book INTEGER, series INTEGER)`,
		`CREATE TABLE data (id INTEGER PRIMARY KEY, book INTEGER, format TEXT)`,
		`CREATE TABLE ratings (id INTEGER PRIMARY KEY, rating INTEGER)`,
		`CREATE TABLE books_ratings_link (book INTEGER, rating INTEGER)`,
		`INSERT INTO authors (id, name) VALUES (1, 'Ursula K. Le Guin')`,
		`INSERT INTO tags (id, name) VALUES (1, 'Science Fiction')`,
		`INSERT INTO ratings (id, rating) VALUES (1, 10)`,
		`INSERT INTO books (id, title, series_index, path, has_cover, pubdate, isbn) VALUES
			(1, 'The Dispossessed', 1, 'Ursula K. Le Guin/The Dispossessed (1)', 1, '1974-01-01', '9780061054884'),
			(2, 'The Left Hand of Darkness', 1, 'Ursula K. Le Guin/The Left Hand of Darkness (2)', 0, '1969-01-01', '')`,
		`INSERT INTO books_authors_link (book, author) VALUES (1, 1), (2, 1)`,
		`INSERT INTO books_tags_link (book, tag) VALUES (1, 1), (2, 1)`,
		`INSERT INTO comments (book, text) VALUES (1, '<p>Anarchist moon.</p>')`,
		`INSERT INTO identifiers (book, type, val) VALUES (1, 'isbn', '9780061054884')`,
		`INSERT INTO data (book, format) VALUES (1, 'EPUB'), (2, 'EPUB')`,
		`INSERT INTO books_ratings_link (book, rating) VALUES (1, 1)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	coverDir := filepath.Join(dir, "Ursula K. Le Guin", "The Dispossessed (1)")
	if err := os.MkdirAll(coverDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(coverDir, "cover.jpg"), []byte("fake-cover"), 0o644)
}
