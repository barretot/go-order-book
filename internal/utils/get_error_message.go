package utils

import "github.com/go-playground/validator/v10"

func GetErrorMessage(fieldErr validator.FieldError) string {
	field := fieldErr.Field()

	switch fieldErr.Tag() {
	case "required":
		return field + " is required."
	case "numeric":
		return field + " must be numeric."
	case "min":
		return field + " cannot be empty."
	case "uuid":
		return field + " must be a valid UUID."
	case "oneof":
		return field + " has an invalid option."
	default:
		return field + " is invalid."

	}
}
