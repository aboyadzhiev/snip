package model

import (
	"context"
	"github.com/go-playground/validator/v10"
)

type Validatable interface {
	Validate(ctx context.Context, validate *validator.Validate) map[string]string
}
