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
			address TEXT NOT NULL,
			redirect INTEGER NOT NULL DEFAULT 0 CHECK (redirect In (0, 1)),
			loop INTEGER NOT NULL DEFAULT 0 CHECK (loop in (0, 1))
		) WITHOUT ROWID;`)

	if err != nil {
		log.Printf("\033[31mwiki_pages table could not be created!\033[0m")
		return nil, err
	}

	return &Storage{db: d}, nil
}

// this performs a db write
func (s *Storage) Insert(id string, link core.ParseRes) error {
	return nil
}

func (s *Storage) Delete(id string) error {
	return nil
}

func (s *Storage) Get(id string) (*core.ParseRes, error) {
	return nil, nil
}
