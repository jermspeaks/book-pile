# Book Pile

A local app for the piles of books you own (physical or digital) and the ones you covet. Scan covers, rate how badly you want to read them, and get a suggestion for what to open next.

Single user, no login. Go API + SQLite on disk + React cover grid.

## Run

You need Go 1.25+ (the toolchain may download 1.26) and Node 22+.

```bash
# API on http://127.0.0.1:8080
go run ./cmd/server

# UI on http://127.0.0.1:5173 (proxies /api and /covers)
cd web && npm install && npm run dev
```

Or `make dev` to start both. Data lives in `./data/` (`bookpile.db` and copied covers). Override with `BOOKPILE_DATA` and `BOOKPILE_ADDR`.

A tiny sample Calibre library is in `testdata/calibre` if you want to try Import without your real library.

## Import Calibre

Open **Import**, paste the Calibre library folder (the one that contains `metadata.db`), preview, then import. Re-import is safe: catalog fields update, but want ratings, read status, physical ownership, and collections stay as you set them.

## Tests

```bash
go test ./...
cd web && npm test
```

SQL queries are generated with [sqlc](https://sqlc.dev) (`sqlc generate`). Migrations run automatically via goose on startup.
