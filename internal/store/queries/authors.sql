-- name: GetAuthor :one
SELECT id, name FROM authors WHERE id = ?;

-- name: GetAuthorByName :one
SELECT id, name FROM authors WHERE name = ?;

-- name: InsertAuthor :one
INSERT INTO authors (name) VALUES (?)
RETURNING id, name;

-- name: ListAuthors :many
SELECT
    a.id,
    a.name,
    COUNT(b.id) AS book_count,
    CAST(COALESCE(SUM(CASE WHEN b.read_status = 'read' THEN 1 ELSE 0 END), 0) AS INTEGER) AS read_count,
    CAST(COALESCE(SUM(CASE WHEN b.read_status = 'unread' THEN COALESCE(b.want_rating, 0) ELSE 0 END), 0) AS INTEGER) AS unread_want_sum,
    CAST(COALESCE(SUM(CASE WHEN b.read_status = 'unread' AND (b.owned_physical = 1 OR b.owned_digital = 1) THEN 1 ELSE 0 END), 0) AS INTEGER) AS unread_owned,
    CAST(COALESCE(SUM(CASE WHEN b.read_status = 'unread' AND b.owned_physical = 0 AND b.owned_digital = 0 THEN 1 ELSE 0 END), 0) AS INTEGER) AS unread_coveting
FROM authors a
LEFT JOIN book_authors ba ON ba.author_id = a.id
LEFT JOIN books b ON b.id = ba.book_id
GROUP BY a.id
ORDER BY unread_want_sum DESC, read_count DESC, a.name COLLATE NOCASE;

-- name: ListBookAuthors :many
SELECT ba.book_id, a.id AS author_id, a.name
FROM book_authors ba
JOIN authors a ON a.id = ba.author_id
ORDER BY a.name COLLATE NOCASE;

-- name: ListAuthorsForBooks :many
SELECT ba.book_id, a.id AS author_id, a.name
FROM book_authors ba
JOIN authors a ON a.id = ba.author_id
WHERE ba.book_id IN (sqlc.slice('book_ids'))
ORDER BY a.name COLLATE NOCASE;

-- name: ListBooksByAuthor :many
SELECT
    b.id, b.title, b.subtitle, b.isbn, b.published_year, b.description, b.cover_path,
    b.series_name, b.series_index, b.owned_physical, b.owned_digital, b.read_status,
    b.want_rating, b.quality_rating, b.calibre_id, b.created_at, b.updated_at
FROM books b
JOIN book_authors ba ON ba.book_id = b.id
WHERE ba.author_id = ?
ORDER BY b.title COLLATE NOCASE;

-- name: DeleteBookAuthors :exec
DELETE FROM book_authors WHERE book_id = ?;

-- name: InsertBookAuthor :exec
INSERT OR IGNORE INTO book_authors (book_id, author_id) VALUES (?, ?);

-- name: ListReadCountByAuthor :many
SELECT ba.author_id, COUNT(*) AS read_count
FROM book_authors ba
JOIN books b ON b.id = ba.book_id
WHERE b.read_status = 'read'
GROUP BY ba.author_id;
