package service

import (
	"context"
	"fmt"
	"github.com/aboyadzhiev/snip/internal/model"
	"github.com/jxskiss/base62"
	"sync/atomic"
)

type URLShortener interface {
	Shorten(ctx context.Context, url string) (string, error)
	Resolve(ctx context.Context, slug string) (string, error)
}

type urlShortener struct {
	hostname string
	sequence atomic.Int64
}

func (u *urlShortener) Shorten(_ context.Context, url string) (string, error) {
	id := u.nextId()
	slug := string(base62.FormatInt(id))
	shortenedURL := &model.ShortenedURL{
		Id:          id,
		Slug:        slug,
		OriginalURL: url,
	}

	return fmt.Sprintf("%s/%s", u.hostname, shortenedURL.Slug), nil
}

func (u *urlShortener) Resolve(_ context.Context, slug string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (u *urlShortener) nextId() int64 {
	return u.sequence.Add(1)
}

func NewURLShortener(hostname string) URLShortener {
	return &urlShortener{
		hostname: hostname,
		sequence: atomic.Int64{},
	}
}
