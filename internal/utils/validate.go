package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var (
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)
	validate   *validator.Validate
)

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})

	_ = validate.RegisterValidation("phone", validatePhone)
	_ = validate.RegisterValidation("password", validatePassword)
	_ = validate.RegisterValidation("date", validateDate)
	_ = validate.RegisterValidation("datetime", validateDateTime)
	_ = validate.RegisterValidation("timeofday", validateTimeOfDay)
	_ = validate.RegisterValidation("softemail", validateSoftEmail)
}

func validatePhone(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return true
	}
	return phoneRegex.MatchString(value)
}

func validateSoftEmail(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return true
	}
	addr, err := mail.ParseAddress(value)
	return err == nil && addr.Address == value
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}

	hasLetter := false
	hasDigit := false
	for _, ch := range password {
		switch {
		case unicode.IsLetter(ch):
			hasLetter = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	return hasLetter && hasDigit
}

func validateDate(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return true
	}
	_, err := time.Parse("2006-01-02", value)
	return err == nil
}

func validateDateTime(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return true
	}
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func validateTimeOfDay(fl validator.FieldLevel) bool {
	value := strings.TrimSpace(fl.Field().String())
	if value == "" {
		return true
	}
	_, err := time.Parse("15:04", value)
	return err == nil
}

// DecodeAndValidate decodes JSON into dst and runs struct validation.
// On failure it writes the error response and returns false.
func DecodeAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		Error(w, http.StatusBadRequest, "Invalid JSON body", err)
		return false
	}
	if sources := ValidateStruct(dst); len(sources) > 0 {
		ValidationError(w, sources)
		return false
	}
	return true
}

// ValidateStruct validates a struct using go-playground/validator tags.
func ValidateStruct(s any) []ErrorSource {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	var sources []ErrorSource
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		sources = append(sources, ErrorSource{Path: "", Message: err.Error()})
		return sources
	}

	for _, fe := range validationErrors {
		sources = append(sources, ErrorSource{
			Path:    fe.Field(),
			Message: validationMessage(fe),
		})
	}
	return sources
}

func validationMessage(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email", "softemail":
		return "Invalid email address"
	case "phone":
		return "Invalid phone number. Use international format e.g. +1234567890"
	case "password":
		return "Password must be at least 6 characters and include a letter and a number"
	case "date":
		return fmt.Sprintf("%s must be YYYY-MM-DD", field)
	case "datetime":
		return fmt.Sprintf("%s must be RFC3339 datetime", field)
	case "timeofday":
		return fmt.Sprintf("%s must be HH:MM (24-hour)", field)
	case "min":
		return fmt.Sprintf("%s cannot be empty", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, fe.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// ParseRFC3339 parses a validated RFC3339 timestamp.
func ParseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(value))
}

// ParseDate parses a validated YYYY-MM-DD date.
func ParseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

// Email keeps a small helper for rare non-struct checks (e.g. headers).
func Email(email string) bool {
	if email == "" {
		return false
	}
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
