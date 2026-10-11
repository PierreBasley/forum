package login

import "regexp"

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	hashDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	hashUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hashSpecial := regexp.MustCompile(`[!@#~$%^&*()+|_.,<>?/{}\-]`).MatchString(password)

	return hashDigit && hashUpper && hashSpecial
}
