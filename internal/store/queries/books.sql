-- name: GetBook :one
SELECT
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at
FROM books
WHERE id = ?;

-- name: GetBookByCalibreID :one
SELECT
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at
FROM books
WHERE calibre_id = ?;

-- name: GetBookByISBN :one
SELECT
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at
FROM books
WHERE isbn = ? AND isbn != ''
LIMIT 1;

-- name: ListBooksByNormalizedTitle :many
SELECT
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at
FROM books
WHERE lower(trim(title)) = lower(trim(?));

-- name: ListBooks :many
SELECT
    b.id, b.title, b.subtitle, b.isbn, b.published_year, b.description, b.cover_path,
    b.series_name, b.series_index, b.owned_physical, b.owned_digital, b.read_status,
    b.want_rating, b.quality_rating, b.calibre_id, b.created_at, b.updated_at
FROM books b
WHERE (sqlc.narg('read_status') IS NULL OR b.read_status = sqlc.narg('read_status'))
  AND (sqlc.narg('min_want') IS NULL OR b.want_rating >= sqlc.narg('min_want'))
  AND (sqlc.narg('pile_id') IS NULL OR EXISTS (
        SELECT 1 FROM book_piles bp WHERE bp.book_id = b.id AND bp.pile_id = sqlc.narg('pile_id')))
  AND (sqlc.narg('collection_id') IS NULL OR EXISTS (
        SELECT 1 FROM book_collections bc WHERE bc.book_id = b.id AND bc.collection_id = sqlc.narg('collection_id')))
  AND (sqlc.narg('genre_id') IS NULL OR EXISTS (
        SELECT 1 FROM book_genres bg WHERE bg.book_id = b.id AND bg.genre_id = sqlc.narg('genre_id')))
  AND (sqlc.narg('author_id') IS NULL OR EXISTS (
        SELECT 1 FROM book_authors ba WHERE ba.book_id = b.id AND ba.author_id = sqlc.narg('author_id')))
  AND (
        sqlc.narg('ownership') IS NULL
        OR (sqlc.narg('ownership') = 'physical' AND b.owned_physical = 1 AND b.owned_digital = 0)
        OR (sqlc.narg('ownership') = 'digital' AND b.owned_digital = 1 AND b.owned_physical = 0)
        OR (sqlc.narg('ownership') = 'both' AND b.owned_physical = 1 AND b.owned_digital = 1)
        OR (sqlc.narg('ownership') = 'coveting' AND b.owned_physical = 0 AND b.owned_digital = 0)
      )
  AND (
        sqlc.narg('q') IS NULL
        OR b.title LIKE '%' || sqlc.narg('q') || '%'
        OR EXISTS (
            SELECT 1 FROM book_authors ba
            JOIN authors a ON a.id = ba.author_id
            WHERE ba.book_id = b.id AND a.name LIKE '%' || sqlc.narg('q') || '%'
        )
      );

-- name: InsertBook :one
INSERT INTO books (
    title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?,
    ?, ?, ?, datetime('now'), datetime('now')
)
RETURNING
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at;

-- name: UpdateCalibreMetadata :one
UPDATE books SET
    title = ?,
    isbn = ?,
    published_year = ?,
    description = ?,
    series_name = ?,
    series_index = ?,
    owned_digital = 1,
    quality_rating = ?,
    calibre_id = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at;

-- name: UpdateBookUserFields :one
UPDATE books SET
    title = ?,
    subtitle = ?,
    isbn = ?,
    published_year = ?,
    description = ?,
    series_name = ?,
    want_rating = ?,
    read_status = ?,
    owned_physical = ?,
    owned_digital = ?,
    updated_at = datetime('now')
WHERE id = ?
RETURNING
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at;

-- name: SetCoverPath :exec
UPDATE books SET cover_path = ?, updated_at = datetime('now') WHERE id = ?;

-- name: ListUnreadBooks :many
SELECT
    id, title, subtitle, isbn, published_year, description, cover_path,
    series_name, series_index, owned_physical, owned_digital, read_status,
    want_rating, quality_rating, calibre_id, created_at, updated_at
FROM books
WHERE read_status = 'unread';
