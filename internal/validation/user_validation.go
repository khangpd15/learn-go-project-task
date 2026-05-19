package validation

import "regexp"

func IsValidPassword(password string) bool {
	if len(password) < 6 {
		return false
	}

	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[@$!%*?&]`).MatchString(password)

	return hasLower && hasUpper && hasNumber && hasSpecial
}

func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	match, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return false
	}
	return match
}

func IsValidIdUser(id int) bool {
	return id > 0
}
