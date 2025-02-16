package store

import (
	"context"
	"errors"
	"github.com/aboyadzhiev/snip/server/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrShortenedURLNotFound = errors.New("shortened url not found")

type ShortenedURL interface {
	Find(ctx context.Context, id int64) (*model.ShortenedURL, error)
	Save(ctx context.Context, shortenedURL *model.ShortenedURL) error
}

type shortenedURLPG struct {
	db *pgxpool.Pool
}

func (s *shortenedURLPG) Find(ctx context.Context, id int64) (*model.ShortenedURL, error) {
	var shortenedURL model.ShortenedURL
	sql := "SELECT id, slug, original_url, created_at FROM url_map WHERE id = $1"
	err := s.db.QueryRow(ctx, sql, id).Scan(&shortenedURL.Id, &shortenedURL.Slug, &shortenedURL.OriginalURL, &shortenedURL.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShortenedURLNotFound
		}
		return nil, err
	}

	return &shortenedURL, nil
}

func (s *shortenedURLPG) Save(ctx context.Context, shortenedURL *model.ShortenedURL) error {
	sql := "INSERT INTO url_map (id, slug, original_url) VALUES ($1, $2, $3)"
	_, err := s.db.Exec(ctx, sql, shortenedURL.Id, shortenedURL.Slug, shortenedURL.OriginalURL)
	if err != nil {
		return err
	}

	return nil
}

func NewShortenedURL(db *pgxpool.Pool) ShortenedURL {
	return &shortenedURLPG{db: db}
}
