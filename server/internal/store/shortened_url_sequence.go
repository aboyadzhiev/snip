package store

import (
	"context"
	"github.com/valkey-io/valkey-go"
)

type ShortenedURLSequence interface {
	NextId(ctx context.Context) (int64, error)
}

const shortenedURLSequenceKey = "ShortenedURLSequence"

type shortenedURLSequenceValkey struct {
	client valkey.Client
}

func (s *shortenedURLSequenceValkey) NextId(ctx context.Context) (int64, error) {
	return s.client.Do(ctx, s.client.B().Incr().Key(shortenedURLSequenceKey).Build()).AsInt64()
}

func NewShortenedURLSequence(client valkey.Client) ShortenedURLSequence {
	return &shortenedURLSequenceValkey{client: client}
}
