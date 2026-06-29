package validator

import (
	"fmt"
	"net/url"

	"github.com/go-playground/validator/v10"
	"pocket-app/internal/pkg/apperror"
)

var validate = validator.New()

func init() {
	// Custom validator for valid URL
	_ = validate.RegisterValidation("valid_url", func(fl validator.FieldLevel) bool {
		u, err := url.ParseRequestURI(fl.Field().String())
		if err != nil {
			return false
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			return false
		}
		return true
	})
}

func Validate(s interface{}) []apperror.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var errors []apperror.FieldError
	for _, err := range err.(validator.ValidationErrors) {
		var message string
		switch err.Tag() {
		case "required":
			message = fmt.Sprintf("%s is required", err.Field())
		case "email":
			message = fmt.Sprintf("%s format is invalid", err.Field())
		case "min":
			message = fmt.Sprintf("%s must be at least %s characters", err.Field(), err.Param())
		case "max":
			message = fmt.Sprintf("%s must not exceed %s characters", err.Field(), err.Param())
		case "oneof":
			message = fmt.Sprintf("%s must be one of [%s]", err.Field(), err.Param())
		case "valid_url":
			message = "URL is invalid"
		default:
			message = fmt.Sprintf("%s is invalid", err.Field())
		}

		errors = append(errors, apperror.FieldError{
			Field:   err.Field(), 
			Message: message,
		})
	}

	return errors
}
