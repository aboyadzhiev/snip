package service

import (
	"context"
	"github.com/aboyadzhiev/snip/server/internal/urlhaus"
	"github.com/valkey-io/valkey-go"
	"log/slog"
	"time"
)

const maliciousURLsKey = "MaliciousURLs"
const maliciousURLsLastUpdatedAtKey = "MaliciousURLsLastUpdatedAt"

type URLGuardian interface {
	SafeURL(ctx context.Context, url string) (bool, error)
	UpdateDB(ctx context.Context) error
}

type urlGuardian struct {
	urlhausClient urlhaus.Client
	valkeyClient  valkey.Client
	logger        *slog.Logger
}

func (u *urlGuardian) SafeURL(ctx context.Context, url string) (bool, error) {
	isMember, err := u.valkeyClient.Do(ctx, u.valkeyClient.B().Sismember().Key(maliciousURLsKey).Member(url).Build()).AsBool()
	if err != nil {
		u.logger.Error("Error while determining whether URL is safe.", "url", url, "err", err)
		return false, err
	}

	return !isMember, nil
}

func (u *urlGuardian) UpdateDB(ctx context.Context) error {
	// Checking the time from the last call to comply with their API requirements. See https://urlhaus.abuse.ch/api/
	exists, err := u.valkeyClient.Do(ctx, u.valkeyClient.B().Exists().Key(maliciousURLsLastUpdatedAtKey).Build()).AsBool()
	if err != nil {
		return err
	}
	if exists {
		lastUpdatedAt, err := u.valkeyClient.Do(ctx, u.valkeyClient.B().Get().Key(maliciousURLsLastUpdatedAtKey).Build()).ToString()
		if err != nil {
			return err
		}
		lastUpdatedAtTime, err := time.Parse(time.RFC3339, lastUpdatedAt)
		if err != nil {
			return err
		}
		diff := time.Since(lastUpdatedAtTime)
		if diff < 5*time.Minute {
			u.logger.WarnContext(ctx, "There are less that 5 minutes from the last update - skipping...")
			return nil
		}
	}

	urls, err := u.urlhausClient.FetchAll(ctx)
	if err != nil {
		return err
	}

	for _, url := range urls {
		u.valkeyClient.Do(ctx, u.valkeyClient.B().Sadd().Key(maliciousURLsKey).Member(url.URL).Build())
	}

	u.valkeyClient.Do(ctx, u.valkeyClient.B().Set().Key(maliciousURLsLastUpdatedAtKey).Value(time.Now().UTC().Format(time.RFC3339)).Build())

	return nil
}

func NewURLGuardian(urlhausClient urlhaus.Client, valkeyClient valkey.Client, logger *slog.Logger) URLGuardian {
	return &urlGuardian{
		urlhausClient: urlhausClient,
		valkeyClient:  valkeyClient,
		logger:        logger,
	}
}
