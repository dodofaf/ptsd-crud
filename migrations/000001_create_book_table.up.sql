CREATE TABLE IF NOT EXISTS book (
    id UUID PRIMARY KEY,
    author TEXT NOT NULL,
    page_count INT NOT NULL CHECK (page_count > 0),
    genre TEXT NOT NULL,
    publication_date DATE NOT NULL
);
