package store

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/roman-duke/philosophy-viz/internal/core"
	_ "modernc.org/sqlite"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(path string) (*Storage, error) {
	d, _ := sql.Open("sqlite", path)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err := d.PingContext(ctx)

	if err != nil {
		log.Printf("\033[31mdb connection could not be opened!\033[0m")
		return nil, err
	}

	_, err = d.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS wiki_pages (
			title TEXT PRIMARY KEY,
			content TEXT NOT NULL
		);`)

	if err != nil {
		log.Printf("\033[31mwiki_pages table could not be created!\033[0m")
		return nil, err
	} else {
		log.Printf("\033[32mwiki_pages table created successfully!\033[0m")
	}

	return &Storage{db: d}, nil
}

// this performs a db write
func (s *Storage) Insert(id string, page core.WikiPageRes) error {
	_, err := s.db.ExecContext(context.Background(), `
	INSERT INTO wiki_pages
	(title, content)
	VALUES (?, ?)
	ON CONFLICT(title) DO NOTHING;
	`, id, page.Parse.Text)

	if err != nil {
		log.Printf("\033[31merror inserting into wiki_pages table\033[0m")
		return err
	}

	return nil
}

func (s *Storage) Delete(id string) error {
	_, err := s.db.ExecContext(context.Background(), `
		DELETE FROM wiki_pages
		WHERE title = ?
	`, id)

	if err != nil {
		log.Printf("\033[31merror deleting from wiki_pages table\033[0m")
		return err
	}

	return nil
}

func (s *Storage) Get(id string) (*core.WikiPageRes, error) {
	var (
		content string
	)

	rowRes := s.db.QueryRowContext(context.Background(), `
		SELECT content FROM wiki_pages
		WHERE title = ?
	`, id)

	err := rowRes.Scan(&content)

	if err != nil {
		return nil, err
	}

	var result core.WikiPageRes
	result.Parse.Text = content

	return &result, nil
}
