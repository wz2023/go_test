package session

import (
	"errors"
	"sync"
)

var GMgr = NewManger()

type Manager struct {
	sync.RWMutex
	us map[string]*Session // user_id<->session
}

func NewManger() *Manager {
	return &Manager{
		us: make(map[string]*Session),
	}
}

func (m *Manager) Bind(userID string, session *Session) {
	m.Lock()
	defer m.Unlock()

	session.Bind(userID)

	m.us[userID] = session
}

func (m *Manager) GetSessionByUserID(userID string) (*Session, error) {
	m.RLock()
	defer m.RUnlock()

	if s, ok := m.us[userID]; ok {
		return s, nil
	}
	return nil, errors.New("not find session")
}
