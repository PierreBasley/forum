package login

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

	/* TODO Créer une fonction qui ouvre une session pour le client
	err := session.CreateSession(sessionID)
	if err != nil {
		log.Print("Could not create session for client", sessionID, ":", err)
	}*/

	expiresAt := time.Now().Add(24 * time.Hour)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.String(),
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false,
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
	/* TODO Créer une fonction DestroySession pour fermer la session du client
	 * côté serveur (libère le mémoire ou l'espace disque associé)
	cookie, err := r.Cookie("session_id")
	if err != nil {
		log.Print("")
	}

	if cookie.Value != "" {
		session.DestroySession()
	}
	*/
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
}
