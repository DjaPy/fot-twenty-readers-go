package telegram

import (
	"sync"
)

type SessionManager struct {
	sessions map[int64]UserSession
	mu       sync.RWMutex
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[int64]UserSession),
	}
}

func (sm *SessionManager) GetSession(userID int64) *UserSession {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[userID]
	if !exists {
		return &UserSession{State: StateIdle}
	}
	return &session
}

func (sm *SessionManager) SetSession(userID int64, session *UserSession) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.sessions[userID] = *session
}

func (sm *SessionManager) DeleteSession(userID int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, userID)
}

func (sm *SessionManager) UpdateState(userID int64, state RegistrationState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[userID]
	if !exists {
		session = UserSession{}
		sm.sessions[userID] = session
	}
	session.State = state
}
