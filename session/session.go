package session

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

const SESSION_DURATION = time.Hour * 3

// Session contient les données d'une connexion authentifiée.
type Session struct {
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Manager conserve les sessions et protège les accès à la map.
type Manager struct {
	mu       sync.RWMutex //agit comme verrou et éviter les data races lorsqu'il y a plusieurs requêtes qui arrivent en même temps
	sessions map[string]Session
}

// NewManager crée un gestionnaire prêt à recevoir des sessions.
func NewManager() *Manager {
	return &Manager{
		sessions: make(map[string]Session),
	}
}

func (m *Manager) CreationSession(userID string) (string, error) {

	if userID == "" {
		return "", errors.New("Erreur : l'userID est vide.")
	}

	now := time.Now()
	s := Session{
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(SESSION_DURATION),
	}

	sessionID, err := uuid.NewV4()
	if err != nil {
		return "", errors.New("Erreur : la génération de l'uuid a échoué.")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sessions[sessionID.String()]; ok == true {
		return "", errors.New("identifiant de session déjà existant")
	}

	m.sessions[sessionID.String()] = s

	return sessionID.String(), nil
}

func (m *Manager) RechercheSession(sessionID string) (Session, bool) {

	if sessionID == "" {
		log.Println("Erreur : la sessionID est vide")
		return Session{}, false
	}

	m.mu.RLock()
	s, exist := m.sessions[sessionID]
	m.mu.RUnlock()

	if exist == false {
		log.Println("Il n'y a pas de session existante")
		return Session{}, exist
	}

	if time.Now().Before(s.ExpiresAt) == false {
		log.Println("La session est expirée")
		return Session{}, false
	}

	return s, true
}

func (m *Manager) SuppressionSession(sessionID string) {
	if sessionID == "" {
		log.Println("Erreur : la sessionID est vide")
		return
	}
	m.mu.Lock()
	delete(m.sessions, sessionID)
	m.mu.Unlock()
}

func (m *Manager) NettoyageSessionsExpirees() {
	now := time.Now()
	m.mu.Lock()
	for id, s := range m.sessions {
		if now.Before(s.ExpiresAt) == false {
			delete(m.sessions, id)
		}
	}
	m.mu.Unlock()
}
