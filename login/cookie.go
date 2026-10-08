package login

import (
	"net/http"
	"time"
)

func CreateCookie(sessionID string, expiresAt time.Time, cookieName string) *http.Cookie {

	cookieUser := &http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	return cookieUser
}

func GetSessionCookie(r *http.Request, cookieName string) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func DeleteSessionCookie(w http.ResponseWriter, cookieName string) {

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
}
