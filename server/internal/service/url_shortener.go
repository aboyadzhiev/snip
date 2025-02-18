package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/aboyadzhiev/snip/server/internal/model"
	"github.com/aboyadzhiev/snip/server/internal/store"
	"github.com/jxskiss/base62"
)

var ErrMaliciousURLDetected = errors.New("malicious URL detected")
var ErrIllegalSlug = errors.New("the given slug doesn't represent base62 encoded integer value")

type URLShortener interface {
	Shorten(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, slug string) (string, error)
}

type urlShortener struct {
	hostname string
	sequence store.ShortenedURLSequence
	store    store.ShortenedURL
	guardian URLGuardian
}

func (s *urlShortener) Shorten(ctx context.Context, url string) (string, error) {
	safeURL, err := s.guardian.SafeURL(ctx, url)
	if err != nil {
		return "", err
	}

	if !safeURL {
		return "", ErrMaliciousURLDetected
	}

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
		return "", ErrIllegalSlug
	}

	shortenedURL, err := s.store.Find(ctx, id)
	if err != nil {
		return "", err
	}

	return shortenedURL.OriginalURL, nil
}

func NewURLShortener(hostname string, sequence store.ShortenedURLSequence, store store.ShortenedURL, guardian URLGuardian) URLShortener {
	return &urlShortener{
		hostname: hostname,
		sequence: sequence,
		store:    store,
		guardian: guardian,
	}
}
