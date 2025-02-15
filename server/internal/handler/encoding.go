package handler

import (
	"encoding/json"
	"fmt"
	"github.com/aboyadzhiev/snip/server/internal/model"
	"github.com/go-playground/validator/v10"
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

func decodeValidatable[T model.Validatable](r *http.Request, validate *validator.Validate) (T, map[string]string, error) {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var v T
	if err := decoder.Decode(&v); err != nil {
		return v, nil, fmt.Errorf("decode json: %w", err)
	}
	if problems := v.Validate(r.Context(), validate); len(problems) > 0 {
		return v, problems, nil
	}
	return v, nil, nil
}
