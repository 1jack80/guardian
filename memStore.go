package guardian

import (
	"errors"
)

// InMemoryStore is an in-memory implementation of the Storer interface.
type InMemoryStore struct {
	data map[string]*Session
}

// NewInMemoryStore creates a new instance of InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]*Session),
	}
}

// get retrieves session data from the in-memory store.
func (s *InMemoryStore) Get(sessionID string) (*Session, error) {
	session, ok := s.data[sessionID]
	if !ok {
		return nil, errors.New("Session not found")
	}
	return session, nil
}

// save saves a session into the in-memory store.
func (s *InMemoryStore) Save(session *Session) error {
	s.data[session.ID] = session
	return nil
}

// delete deletes session data from the in-memory store.
func (s *InMemoryStore) Delete(sessionID string) error {

	delete(s.data, sessionID)
	return nil
}

// Update updates session data in the in-memory store.
func (s *InMemoryStore) Update(sessionID string, newSession *Session) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[sessionID] = newSession
	return nil
}
