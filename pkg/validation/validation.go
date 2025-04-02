package validation

import "regexp"

// IsValidEmail validates email format using a regular expression
func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

// IsValidUUID checks if a string is a valid UUID.
func IsValidUUID(uuid string) bool {
	var uuidRegex = `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	re := regexp.MustCompile(uuidRegex)
	return re.MatchString(uuid)
}

// IsValidPassword checks if the password meets the strength requirements
func IsValidPassword(password string) bool {
	// Check password length
	if len(password) < 8 {
		return false
	}

	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one number
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char == '@' || char == '#' || char == '$' || char == '%' || char == '&' || char == '*' || char == '!' || char == '^':
			hasSpecial = true
		}
	}

	// Password must contain at least one lowercase letter, one uppercase letter, one digit, and one special character
	return hasLower && hasUpper && hasDigit && hasSpecial
}
