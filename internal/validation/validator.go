package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type RequestValidator struct {
	validate *validator.Validate
}

func NewValidator() *RequestValidator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := field.Tag.Get("json")
		if name == "" {
			name = field.Tag.Get("form")
		}
		if name == "" {
			return field.Name
		}
		parts := strings.SplitN(name, ",", 2)
		if parts[0] == "-" || parts[0] == "" {
			return field.Name
		}
		return parts[0]
	})

	if err := v.RegisterValidation("uuid4", func(fl validator.FieldLevel) bool {
		value := strings.TrimSpace(fl.Field().String())
		if value == "" {
			return true
		}
		_, err := uuid.Parse(value)
		return err == nil
	}); err != nil {
		// Log the error but continue - validation will still work without custom uuid4 validator
		fmt.Printf("WARNING: Failed to register uuid4 validation: %v\n", err)
	}

	return &RequestValidator{validate: v}
}

func (v *RequestValidator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

func (v *RequestValidator) Engine() interface{} {
	return v.validate
}

func ExtractErrors(err error) map[string]string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}

	issues := make(map[string]string, len(ve))
	for _, fe := range ve {
		field := fe.Field()
		if field == "" {
			field = fe.StructField()
		}
		issues[field] = messageForFieldError(fe)
	}
	return issues
}

func messageForFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters", fe.Param())
	case "len":
		return fmt.Sprintf("Must be %s characters", fe.Param())
	case "uuid4":
		return "Must be a valid UUID"
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "eqfield":
		target := strings.ToLower(strings.ReplaceAll(fe.Param(), "_", " "))
		return fmt.Sprintf("Must match %s", target)
	default:
		return "Invalid value"
	}
}
