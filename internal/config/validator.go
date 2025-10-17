package config

import (
	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

// newValidator creates a new validator instance with custom validation rules.
func newValidator() *validator.Validate {
	return validator.New()
}

// Validate validates the configuration struct using go-playground/validator.
// It returns a detailed error message if validation fails.
func Validate(cfg *Config) error {
	return validate.Struct(cfg)
}
