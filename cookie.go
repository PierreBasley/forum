package main

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid"
)

func CreateSessionCookie(w http.ResponseWriter) (string, error) {

	sessionID, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	}

	http.SetCookie(w, cookie)

	return sessionID.String(), nil
}

func GetSessionCookie(r *http.Request) (string, error) {

	cookie, err := r.Cookie("session_id")
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func DeleteSessionCookie(w http.ResponseWriter) {

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
}
