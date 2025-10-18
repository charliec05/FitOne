package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomString generates a random string of the specified length
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// IsValidEmail performs basic email validation
func IsValidEmail(email string) bool {
	// Basic email validation - in production, use a proper email validation library
	if len(email) < 5 {
		return false
	}
	
	hasAt := false
	hasDot := false
	
	for i, char := range email {
		if char == '@' {
			if hasAt || i == 0 || i == len(email)-1 {
				return false
			}
			hasAt = true
		}
		if char == '.' && hasAt {
			hasDot = true
		}
	}
	
	return hasAt && hasDot
}
