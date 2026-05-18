package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Validator wraps go-playground/validator with custom rules.
type Validator struct {
	validate *validator.Validate
}

// New creates a Validator instance with custom rules registered.
func New() *Validator {
	v := validator.New()

	_ = v.RegisterValidation("uuid", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if val == "" {
			return false
		}
		_, err := uuid.Parse(val)
		return err == nil
	})

	return &Validator{validate: v}
}

// ValidateStruct validates a struct and returns a formatted error.
func (v *Validator) ValidateStruct(s interface{}) error {
	err := v.validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return fmt.Errorf("validation failed: %w", err)
	}

	var messages []string
	for _, e := range validationErrs {
		messages = append(messages, formatFieldError(e))
	}

	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

func formatFieldError(e validator.FieldError) string {
	field := e.Field()
	tag := e.Tag()
	param := e.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, param)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, param)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	default:
		if param != "" {
			return fmt.Sprintf("%s failed %s(%s) validation", field, tag, param)
		}
		return fmt.Sprintf("%s failed %s validation", field, tag)
	}
}
