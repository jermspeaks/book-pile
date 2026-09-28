CREATE TABLE books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    subtitle TEXT,
    isbn TEXT,
    published_year INTEGER,
    description TEXT,
    cover_path TEXT,
    series_name TEXT,
    series_index REAL,
    owned_physical INTEGER NOT NULL DEFAULT 0,
    owned_digital INTEGER NOT NULL DEFAULT 0,
    read_status TEXT NOT NULL DEFAULT 'unread',
    want_rating INTEGER,
    quality_rating REAL,
    calibre_id INTEGER,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX idx_books_calibre_id ON books (calibre_id) WHERE calibre_id IS NOT NULL;
CREATE INDEX idx_books_isbn ON books (isbn);
CREATE INDEX idx_books_title ON books (title);
CREATE INDEX idx_books_read_status ON books (read_status);
CREATE INDEX idx_books_want_rating ON books (want_rating);

CREATE TABLE authors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE book_authors (
    book_id INTEGER NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    author_id INTEGER NOT NULL REFERENCES authors (id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);

CREATE TABLE piles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE book_piles (
    book_id INTEGER NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    pile_id INTEGER NOT NULL REFERENCES piles (id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, pile_id)
);

CREATE TABLE collections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE book_collections (
    book_id INTEGER NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    collection_id INTEGER NOT NULL REFERENCES collections (id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, collection_id)
);

CREATE TABLE genres (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE book_genres (
    book_id INTEGER NOT NULL REFERENCES books (id) ON DELETE CASCADE,
    genre_id INTEGER NOT NULL REFERENCES genres (id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, genre_id)
);
