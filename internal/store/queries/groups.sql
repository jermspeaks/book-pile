-- name: ListPiles :many
SELECT id, name FROM piles ORDER BY name COLLATE NOCASE;

-- name: GetPile :one
SELECT id, name FROM piles WHERE id = ?;

-- name: GetPileByName :one
SELECT id, name FROM piles WHERE name = ?;

-- name: InsertPile :one
INSERT INTO piles (name) VALUES (?)
RETURNING id, name;

-- name: ListPilesForBook :many
SELECT p.id, p.name
FROM piles p
JOIN book_piles bp ON bp.pile_id = p.id
WHERE bp.book_id = ?
ORDER BY p.name COLLATE NOCASE;

-- name: DeleteBookPiles :exec
DELETE FROM book_piles WHERE book_id = ?;

-- name: InsertBookPile :exec
INSERT OR IGNORE INTO book_piles (book_id, pile_id) VALUES (?, ?);

-- name: ListCollections :many
SELECT id, name FROM collections ORDER BY name COLLATE NOCASE;

-- name: GetCollection :one
SELECT id, name FROM collections WHERE id = ?;

-- name: GetCollectionByName :one
SELECT id, name FROM collections WHERE name = ?;

-- name: InsertCollection :one
INSERT INTO collections (name) VALUES (?)
RETURNING id, name;

-- name: ListCollectionsForBook :many
SELECT c.id, c.name
FROM collections c
JOIN book_collections bc ON bc.collection_id = c.id
WHERE bc.book_id = ?
ORDER BY c.name COLLATE NOCASE;

-- name: DeleteBookCollections :exec
DELETE FROM book_collections WHERE book_id = ?;

-- name: InsertBookCollection :exec
INSERT OR IGNORE INTO book_collections (book_id, collection_id) VALUES (?, ?);

-- name: ListGenres :many
SELECT id, name FROM genres ORDER BY name COLLATE NOCASE;

-- name: GetGenre :one
SELECT id, name FROM genres WHERE id = ?;

-- name: GetGenreByName :one
SELECT id, name FROM genres WHERE name = ?;

-- name: InsertGenre :one
INSERT INTO genres (name) VALUES (?)
RETURNING id, name;

-- name: ListGenresForBook :many
SELECT g.id, g.name
FROM genres g
JOIN book_genres bg ON bg.genre_id = g.id
WHERE bg.book_id = ?
ORDER BY g.name COLLATE NOCASE;

-- name: DeleteBookGenres :exec
DELETE FROM book_genres WHERE book_id = ?;

-- name: InsertBookGenre :exec
INSERT OR IGNORE INTO book_genres (book_id, genre_id) VALUES (?, ?);
