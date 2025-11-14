package handlers

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	"github.com/jamesc159/monmetrics/internal/models"
)

// normalizeEmail converts email to lowercase and trims whitespace
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// sanitizeName removes dangerous characters and normalizes whitespace
func sanitizeName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Remove control characters and normalize spaces
	var result strings.Builder
	prevSpace := false
	for _, r := range name {
		if unicode.IsControl(r) {
			continue
		}
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		result.WriteRune(r)
		prevSpace = false
	}

	return result.String()
}

// validateEmail validates email format using regex
func validateEmail(email string) bool {
	// RFC 5322 simplified email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return emailRegex.MatchString(email) && len(email) <= 254
}

// validatePassword checks password strength according to OWASP guidelines
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	// Check for at least one uppercase, lowercase, digit, and special character
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	return nil
}

// validateName checks if name is valid
func validateName(name string) error {
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}

	if len(name) > 50 {
		return fmt.Errorf("name must not exceed 50 characters")
	}

	// Check for at least one letter
	hasLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}

	if !hasLetter {
		return fmt.Errorf("name must contain at least one letter")
	}

	return nil
}

// validateRegisterRequest validates and normalizes registration request data
func (h *Handlers) validateRegisterRequest(req *models.RegisterRequest) error {
	// Normalize inputs
	req.Email = normalizeEmail(req.Email)
	req.FirstName = sanitizeName(req.FirstName)
	req.LastName = sanitizeName(req.LastName)

	// Validate all fields are present
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("all fields are required")
	}

	// Validate email format
	if !validateEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		return err
	}

	// Validate names
	if err := validateName(req.FirstName); err != nil {
		return fmt.Errorf("invalid first name: %v", err)
	}

	if err := validateName(req.LastName); err != nil {
		return fmt.Errorf("invalid last name: %v", err)
	}

	return nil
}

// hashPassword hashes a password using bcrypt with cost factor 12
func (h *Handlers) hashPassword(password string) (string, error) {
	// Use bcrypt with cost factor 12 (OWASP recommended minimum is 10)
	// Cost 12 provides good balance between security and performance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// verifyPassword verifies a password against a bcrypt hash
func (h *Handlers) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
