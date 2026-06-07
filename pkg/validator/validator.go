package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Use JSON tag names for validation errors instead of struct field names
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidateStruct validates a struct and returns a map of field validation errors.
// If no validation errors exist, it returns nil.
func ValidateStruct(s interface{}) map[string]string {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var valErrors validator.ValidationErrors
	if errors.As(err, &valErrors) {
		errMap := make(map[string]string)
		for _, e := range valErrors {
			errMap[e.Field()] = formatError(e)
		}
		return errMap
	}

	return map[string]string{"_error": err.Error()}
}

// formatError converts a ValidationErrors entry into a readable message.
func formatError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must be at least %s characters long", err.Param())
		}
		return fmt.Sprintf("Must be at least %s", err.Param())
	case "max":
		if err.Type().Kind() == reflect.String {
			return fmt.Sprintf("Must not exceed %s characters", err.Param())
		}
		return fmt.Sprintf("Must not exceed %s", err.Param())
	case "uuid":
		return "Must be a valid UUID"
	default:
		return fmt.Sprintf("Validation failed on rule '%s'", err.Tag())
	}
}
