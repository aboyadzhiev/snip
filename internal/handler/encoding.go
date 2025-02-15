package handler

import (
	"encoding/json"
	"net/http"
)

func encode[T any](w http.ResponseWriter, status int, v T, headers http.Header) error {
	bytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(bytes)

	return nil
}
