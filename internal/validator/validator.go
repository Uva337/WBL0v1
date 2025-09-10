package validator

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() (*Validator, error) {
	v := validator.New()
	return &Validator{validate: v}, nil
}

func (v *Validator) Struct(s any) error {
	err := v.validate.Struct(s)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}