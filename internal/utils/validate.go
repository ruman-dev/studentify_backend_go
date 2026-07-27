package utils

import (
	"net/mail"
	"regexp"
	"unicode"
)

var phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{7,14}$`)

func Email(email string) bool {
	if email == "" {
		return false
	}
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func Phone(phone string) bool {
	return phoneRegex.MatchString(phone)
}

// Password requires at least 6 characters, one letter, and one digit.
func Password(password string) bool {
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
