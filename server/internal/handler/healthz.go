package handler

import (
	"net/http"
)

func Healthz() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload := make(map[string]string)
		payload["status"] = "ok"
		if err := encode[map[string]string](w, http.StatusOK, payload, nil); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
