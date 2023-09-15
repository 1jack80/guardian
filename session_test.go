package guardian_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"

	"github.com/1jack80/guardian"
)

// MockStorage is a mock implementation of the Storer interface for testing.
type MockStorage struct {
	data map[string]guardian.Session
	mu   sync.RWMutex
}

// NewMockStorage creates a new instance of MockStorage.
func NewMockStorage() *MockStorage {
	return &MockStorage{
		data: make(map[string]guardian.Session),
	}
}

// get retrieves session data from the mock storage.
func (s *MockStorage) Get(sessionID string) (guardian.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.data[sessionID]
	if !ok {
		return guardian.Session{}, errors.New("Session not found")
	}
	return session, nil
}

// save saves a session into the mock storage.
func (s *MockStorage) Save(session guardian.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[session.ID] = session
	return nil
}

// delete deletes session data from the mock storage.
func (s *MockStorage) Delete(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, sessionID)
	return nil
}

// Update updates session data in the mock storage.
func (s *MockStorage) Update(sessionID string, newSession guardian.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sessionID == newSession.ID {
		s.data[sessionID] = newSession
		return nil
	} else {
		delete(s.data, sessionID)
		s.data[newSession.ID] = newSession
		return nil
	}
}

func TestValidateNamespace(t *testing.T) {
	err := guardian.ValidateNamespace("one")
	if err != nil {
		t.Error(err)
	}
	err = guardian.ValidateNamespace("one")
	if err == nil {
		t.Error(err)
	}
}

var store = guardian.NewInMemoryStore()
var manager, manager_err = guardian.NewManager("test_manager", store)

// TestSessionManager_CreateSession tests the creation of a new session and validates its attributes.
func TestSessionManager_CreateSession(t *testing.T) {
	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	if manager.ContextKey() == "" {
		t.Fatal("session manager context key is empty")
	}
	if manager.SaveSession(guardian.Session{ID: "one"}) != nil {
		t.Fatalf("manager cannot save session")
	}
}

// TestSessionManager_WatchTimeouts tests the session timeout handling, including renewal and invalidation.

// TestSessionManager_PopulateRequestContext ensures that session data is correctly populated in the request context.
func TestSessionManager_PopulateRequestContext(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	type key string
	ctx := context.Background()
	ctx = context.WithValue(ctx, key("foo"), "bar")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "github.com", nil)
	if err != nil {
		t.Fatalf("unable to create request: %v", err)
	}
	session := manager.CreateSession()

	req = manager.PopulateRequestContext(req, session)

	if req.Context().Value(key("foo")) != "bar" {
		t.Fatalf("existing values in the context were altered")
	}
	if req.Context().Value(manager.ContextKey()) == nil {
		t.Fatal("session was not saved in context under context key")
	}
}

// TestSessionManager_InvalidateSession tests the invalidation of a session and verifies that the session data is removed from the store and the session cookie is expired.
func TestSessionManager_InvalidateSession(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	session := manager.CreateSession()
	sessionID := session.ID
	session.Status = guardian.VALID

	manager.InvalidateSession(sessionID)

	session, err := manager.GetSession(sessionID)
	if err != nil {
		t.Fatalf("unable to get session from manager")
	}
	if session.Status != guardian.INVALID {
		t.Fatalf("session status is not invalid")
	}
}

// Tests for Objectives Not Yet Implemented:

// TestSessionManager_RenewSession tests the session renewal functionality once implemented.
func TestSessionManager_RenewSession(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	session := manager.CreateSession()
	sessionID := session.ID

	session, err := manager.RenewSession(sessionID)
	if err != nil {
		t.Error(err)
	}

	if sessionID == session.ID {
		t.Fatal("session id was not renewed")
	}
	if _, err := manager.GetSession(sessionID); err == nil {
		t.Fatal("old session id not updated in the store")
	}
	if _, err := manager.GetSession(session.ID); err != nil {
		t.Fatal("new sesson id was not added to the store")
	}
}

/*
// TestSessionManager_Customization tests setting custom values for idle timeout, lifetime, and renewal timeout.
func TestSessionManager_Customization(t *testing.T) {
	// Test logic here
}

// TestSessionManager_Concurrency tests concurrent access to the session manager to ensure data consistency.
func TestSessionManager_Concurrency(t *testing.T) {
	// Test logic here
}

// TestSessionManager_ErrorHandling adds tests for error cases and ensures that appropriate error messages are returned.
func TestSessionManager_ErrorHandling(t *testing.T) {
	// Test logic here
}

// TestSessionManager_Logging tests if log messages are generated correctly.
func TestSessionManager_Logging(t *testing.T) {
	// Test logic here
}

// Integration Tests: Consider integration tests where you simulate multiple sessions and interactions between them to ensure the session manager behaves correctly in a real-world scenario.

// Edge Cases: Test edge cases such as when session creation or timeout handling is triggered under various conditions (e.g., minimum timeout values, maximum session data size).

// TestNamespaceManager: If you have not yet implemented multiple session namespaces, create tests to ensure that multiple namespaces can coexist and function correctly.


*/
