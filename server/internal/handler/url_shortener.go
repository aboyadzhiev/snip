package handler

import (
	"errors"
	"github.com/aboyadzhiev/snip/server/internal/model"
	"github.com/aboyadzhiev/snip/server/internal/service"
	"github.com/aboyadzhiev/snip/server/internal/store"
	"github.com/go-playground/validator/v10"
	"net/http"
)

func ShortenURL(shortener service.URLShortener, v *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		shortenURLReq, problems, err := decodeValidatable[model.ShortenURLReq](r, v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(problems) > 0 {
			if err = encode[map[string]string](w, http.StatusBadRequest, problems, nil); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		slug, err := shortener.Shorten(ctx, shortenURLReq.URL)
		if err != nil {
			if errors.Is(err, service.ErrMaliciousURLDetected) {
				w.WriteHeader(http.StatusNotAcceptable)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		payload := model.ShortenURLRes{ShortenURL: slug}
		if err = encode[model.ShortenURLRes](w, http.StatusCreated, payload, nil); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func Resolve(shortener service.URLShortener) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		slug := r.PathValue("slug")
		url, err := shortener.Resolve(ctx, slug)
		if err != nil {
			if errors.Is(err, store.ErrShortenedURLNotFound) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
				return
			}

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, url, http.StatusFound)
	}
}
