package model

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"time"
)

type Validatable interface {
	Validate(ctx context.Context, validate *validator.Validate) map[string]string
}

type ShortenURLReq struct {
	URL string `json:"url" validate:"required,min=16,max=4096,http_url"`
}

func (s ShortenURLReq) Validate(ctx context.Context, validate *validator.Validate) map[string]string {
	err := validate.StructCtx(ctx, s)
	if err != nil {
		// this check is only needed when the code could produce an invalid value
		// for validation such as interface with nil value.
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			panic(err)
		}

		problems := map[string]string{}
		for _, fieldError := range err.(validator.ValidationErrors) {
			problems[fieldError.Field()] = errMessage(fieldError)
		}

		return problems
	}

	return nil

}

func errMessage(err validator.FieldError) string {
	var message string
	switch err.Tag() {
	case "required":
		message = fmt.Sprintf("The '%s' field is required.", err.Field())
	case "min":
		message = fmt.Sprintf("The '%s' field must be at least %s characters long.", err.Field(), err.Param())
	case "max":
		message = fmt.Sprintf("The '%s' field cannot exceed %s characters.", err.Field(), err.Param())
	case "email":
		message = fmt.Sprintf("The '%s' field must be a valid email address.", err.Field())
	case "gte":
		message = fmt.Sprintf("The '%s' must be greater than or equal to %s.", err.Field(), err.Param())
	case "lte":
		message = fmt.Sprintf("The '%s' must be less than or equal to %s.", err.Field(), err.Param())
	case "http_url":
		message = fmt.Sprintf("The '%s' must be valid http(s) URL.", err.Field())
	default:
		message = err.Error() // Fallback to default message
	}

	return message

}

type ShortenURLRes struct {
	ShortenURL string `json:"shortenURL"`
}

type ShortenedURL struct {
	Id          int64
	Slug        string
	OriginalURL string
	CreatedAt   time.Time
}
