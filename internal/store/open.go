package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jermspeaks/book-pile/migrations"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

type DB struct {
	SQL *sql.DB
	Q   *Queries
}

func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(filepath.Join(dataDir, "covers"), 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)",
		filepath.ToSlash(filepath.Join(dataDir, "bookpile.db")))
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("goose dialect: %w", err)
	}
	if err := goose.Up(sqlDB, "."); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return &DB{SQL: sqlDB, Q: New(sqlDB)}, nil
}

func (d *DB) Close() error {
	if d == nil || d.SQL == nil {
		return nil
	}
	return d.SQL.Close()
}

func Ownership(physical, digital int64) string {
	switch {
	case physical != 0 && digital != 0:
		return "both"
	case physical != 0:
		return "physical"
	case digital != 0:
		return "digital"
	default:
		return "coveting"
	}
}

func BoolToInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}

func CoverURL(coverPath *string) *string {
	if coverPath == nil || *coverPath == "" {
		return nil
	}
	url := "/covers/" + filepath.Base(*coverPath)
	return &url
}

func Normalize(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(s))), " ")
}

func StrPtr(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	return &s
}

func EnsureAuthor(ctx context.Context, q *Queries, name string) (Author, error) {
	name = strings.TrimSpace(name)
	a, err := q.GetAuthorByName(ctx, name)
	if err == nil {
		return a, nil
	}
	if err != sql.ErrNoRows {
		return Author{}, err
	}
	return q.InsertAuthor(ctx, name)
}

func EnsurePile(ctx context.Context, q *Queries, name string) (Pile, error) {
	p, err := q.GetPileByName(ctx, name)
	if err == nil {
		return p, nil
	}
	if err != sql.ErrNoRows {
		return Pile{}, err
	}
	return q.InsertPile(ctx, name)
}

func EnsureCollection(ctx context.Context, q *Queries, name string) (Collection, error) {
	c, err := q.GetCollectionByName(ctx, name)
	if err == nil {
		return c, nil
	}
	if err != sql.ErrNoRows {
		return Collection{}, err
	}
	return q.InsertCollection(ctx, name)
}

func EnsureGenre(ctx context.Context, q *Queries, name string) (Genre, error) {
	g, err := q.GetGenreByName(ctx, name)
	if err == nil {
		return g, nil
	}
	if err != sql.ErrNoRows {
		return Genre{}, err
	}
	return q.InsertGenre(ctx, name)
}

func AsInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case []byte:
		var x int64
		_, _ = fmt.Sscan(string(n), &x)
		return x
	default:
		return 0
	}
}
