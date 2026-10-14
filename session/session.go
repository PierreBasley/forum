package session

import (
	"encoding/json"
	"errors"
	"forum/login"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

const SESSION_DURATION = time.Hour * 3

// Session type which contain data needed for a proper authentification.
type Session struct {
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Manager type which manage the sessions and protect against concurrency race.
type Manager struct {
	mu  sync.RWMutex //coordoninate the RW operation on the files.
	dir string
}

type Authentified struct {
	IsActive bool
	Session  Session
}

// NewManager create a manager ready to manage session and files operation.
func NewManager(sessionDir string) (*Manager, error) {
	if sessionDir == "" {
		return nil, errors.New("Le chemin du dossier est vide")
	}

	err := os.MkdirAll(sessionDir, 0o700) //the unix permission 0o700 only allows the owner to read,
	//write and navigate through the repertories created.
	if err != nil {
		return nil, err
	}
	manager := &Manager{dir: sessionDir}
	return manager, nil
}

func (m *Manager) CreationSession(userID string) (string, time.Time, error) {

	if userID == "" {
		return "", time.Time{}, errors.New("Erreur : l'userID est vide.")
	}

	now := time.Now()
	expiresAt := now.Add(SESSION_DURATION)
	s := Session{
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
	}

	sessionID, err := uuid.NewV4()
	if err != nil {
		return "", time.Time{}, errors.New("Erreur : la génération de l'uuid a échoué.")
	}

	data, err := json.Marshal(s)
	if err != nil {
		return "", time.Time{}, err
	}

	sessionPath := filepath.Join(m.dir, sessionID.String()+".json")

	//this lock will allow one execution at the time for every file manipulation.
	m.mu.Lock()
	defer m.mu.Unlock()

	//To explain, we create a file which will contain the json data on a session ID
	//sessionPath is the path of where the file will be created.
	//the flags os.O_WRONLY (open the file in readonly),
	//os.O_CREATE (create a file if it doesn't exist),
	//os.O_EXCL (used with O_CREATE, file must not exist),
	//0o600 is a unix permission allowing read and write a file.
	file, err := os.OpenFile(sessionPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", time.Time{}, err
	}

	_, writeErr := file.Write(data)
	if writeErr != nil {
		closeErr := file.Close()
		removeErr := os.Remove(sessionPath)

		return "", time.Time{}, errors.Join(writeErr, closeErr, removeErr)
	}

	closeErr := file.Close()
	if closeErr != nil {
		removeErr := os.Remove(sessionPath)

		return "", time.Time{}, errors.Join(closeErr, removeErr)
	}
	return sessionID.String(), expiresAt, nil
}

func (m *Manager) RechercheSession(sessionID string) (Session, bool, error) {

	//AI told me to use this part of the code to guarantee that the name will be conform to the format
	//to the creation of the file, rather than using sessionID directly (because of match case ?)
	parsedID, err := uuid.FromString(sessionID)
	if err != nil {
		return Session{}, false, nil
	}

	sessionPath := filepath.Join(m.dir, parsedID.String()+".json")

	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := os.ReadFile(sessionPath)
	if errors.Is(err, os.ErrNotExist) { //err in the case where the file is not found
		return Session{}, false, nil
	}
	if err != nil { //err in other cases of reading files
		return Session{}, false, err
	}

	var s Session

	err = json.Unmarshal(data, &s)
	if err != nil {
		return Session{}, false, err
	}

	if s.UserID == "" || s.CreatedAt.IsZero() == true || s.ExpiresAt.IsZero() == true {
		return Session{}, false, errors.New("données de session incomplètes")
	}

	if s.ExpiresAt.After(s.CreatedAt) == false {
		return Session{}, false, errors.New("expiration de session incohérente")
	}

	if time.Now().Before(s.ExpiresAt) == false {
		return Session{}, false, nil
	}
	return s, true, nil
}

func (m *Manager) IsSessionActive(r *http.Request, cookieName string) (Session, bool, error) {
	sessionID, err := login.GetSessionCookie(r, cookieName)
	if err != nil {
		return Session{}, false, nil
	}

	session, ok, err := m.RechercheSession(sessionID)
	if err != nil {
		return Session{}, false, err
	}
	if ok == false {
		return Session{}, false, nil
	}

	return session, true, nil
}

func (m *Manager) SuppressionSession(sessionID string) error {

	parsedID, err := uuid.FromString(sessionID)
	if err != nil {
		return nil
	}

	sessionPath := filepath.Join(m.dir, parsedID.String()+".json")
	m.mu.Lock()
	defer m.mu.Unlock()
	err = os.Remove(sessionPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	return nil
}

func (m *Manager) NettoyageSessionsExpirees() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	files, err := os.ReadDir(m.dir)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, file := range files {
		if file.Type().IsRegular() == true {
			sessionPathFile := filepath.Join(m.dir, file.Name())
			if filepath.Ext(sessionPathFile) == ".json" {

				uuidText := strings.TrimSuffix(file.Name(), ".json")

				parsedID, err := uuid.FromString(uuidText)
				if err != nil {
					continue
				}
				//This recommandation from AI is linked with the format chosen in the function CreateSession and the uuid format of v4. I think it's kind of useless.
				// if file.Name() != parsedID.String()+".json" {
				// 	continue
				// }
				sessionPath := filepath.Join(m.dir, parsedID.String()+".json")

				data, err := os.ReadFile(sessionPath)
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				if err != nil {
					return err
				}

				var s Session

				err = json.Unmarshal(data, &s)
				if err != nil {
					return err
				}

				if s.UserID == "" || s.CreatedAt.IsZero() == true || s.ExpiresAt.IsZero() == true {
					return errors.New("données de session incomplètes")
				}

				if s.ExpiresAt.After(s.CreatedAt) == false {
					return errors.New("expiration de session incohérente")
				}

				if now.Before(s.ExpiresAt) == false {
					err := os.Remove(sessionPath)
					if err != nil && !errors.Is(err, os.ErrNotExist) {
						return err
					}
				}
			} else {
				continue
			}
		}
	}
	return nil
}
