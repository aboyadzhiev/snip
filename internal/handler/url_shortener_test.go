package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aboyadzhiev/snip/internal/model"
	"github.com/go-playground/validator/v10"
	"net/http"
	"net/http/httptest"
	"testing"
)

type stubURLShortener struct {
	shortenCalls int
}

func (s *stubURLShortener) Shorten(_ context.Context, url string) (string, error) {
	s.shortenCalls++
	return "https://www.snap.it/abcd", nil
}

func (s *stubURLShortener) Resolve(_ context.Context, slug string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func TestShortenURL(t *testing.T) {
	validate := validator.New(validator.WithRequiredStructEnabled())
	t.Run("shorten valid http url", func(t *testing.T) {
		shortenerStub := &stubURLShortener{
			shortenCalls: 0,
		}
		shortenURLReq := &model.ShortenURLReq{
			URL: "https://www.fsf.org/blogs/community/i-love-free-software-2025",
		}
		payload, err := json.Marshal(shortenURLReq)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/shortened-url", bytes.NewBuffer(payload))
		res := httptest.NewRecorder()

		ShortenURL(shortenerStub, validate).ServeHTTP(res, req)

		if res.Code != http.StatusCreated {
			t.Errorf("got %d, want %d", res.Code, http.StatusCreated)
		}

		if shortenerStub.shortenCalls != 1 {
			t.Errorf("got %d shortenCalls, want %d shortenCalls", shortenerStub.shortenCalls, 1)
		}
	})

	t.Run("shorten invalid http url", func(t *testing.T) {
		shortenerStub := &stubURLShortener{
			shortenCalls: 0,
		}
		shortenURLReq := &model.ShortenURLReq{
			URL: "ftp://ftp.example.com",
		}
		payload, err := json.Marshal(shortenURLReq)
		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/api/v1/shortened-url", bytes.NewBuffer(payload))
		res := httptest.NewRecorder()

		ShortenURL(shortenerStub, validate).ServeHTTP(res, req)

		if res.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", res.Code, http.StatusBadRequest)
		}

		if shortenerStub.shortenCalls != 0 {
			t.Errorf("got %d shortenCalls, want %d shortenCalls", shortenerStub.shortenCalls, 0)
		}
	})
}
