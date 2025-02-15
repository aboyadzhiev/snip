package service

import (
	"context"
	"fmt"
	"github.com/aboyadzhiev/snip/server/internal/model"
	"github.com/aboyadzhiev/snip/server/internal/store"
	"github.com/jxskiss/base62"
)

type URLShortener interface {
	Shorten(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, slug string) (string, error)
}

type urlShortener struct {
	hostname string
	sequence store.ShortenedURLSequence
	store    store.ShortenedURL
}

func (s *urlShortener) Shorten(ctx context.Context, url string) (string, error) {
	id, err := s.sequence.NextId(ctx)
	if err != nil {
		return "", err
	}
	slug := string(base62.FormatInt(id))
	shortenedURL := &model.ShortenedURL{
		Id:          id,
		Slug:        slug,
		OriginalURL: url,
	}

	err = s.store.Save(ctx, shortenedURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.hostname, shortenedURL.Slug), nil
}

func (s *urlShortener) Resolve(ctx context.Context, slug string) (string, error) {
	id, err := base62.ParseInt([]byte(slug))
	if err != nil {
		return "", err
	}

	shortenedURL, err := s.store.Find(ctx, id)
	if err != nil {
		return "", err
	}

	return shortenedURL.OriginalURL, nil

}

func NewURLShortener(hostname string, sequence store.ShortenedURLSequence, store store.ShortenedURL) URLShortener {
	return &urlShortener{
		hostname: hostname,
		sequence: sequence,
		store:    store,
	}
}
