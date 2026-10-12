package session

import (
	"forum/models"
)

type Authentified struct {
	IsActive bool
	Session  Session
	User     models.User
}
