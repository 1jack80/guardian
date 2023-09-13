package guardian_test

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"sync"
	"testing"
	"time"

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
func (s *MockStorage) Get(sessionID string) (*guardian.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.data[sessionID]
	if !ok {
		return nil, errors.New("Session not found")
	}
	return &session, nil
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

var manager, manager_err = guardian.NewSessionManager("test_manager", NewMockStorage())

// TestSessionManager_CreateSession tests the creation of a new session and validates its attributes.
func TestSessionManager_CreateSession(t *testing.T) {
	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	if manager.ContextKey() == "" {
		t.Fatal("session manager context key is empty")
	}
	if manager.IdleTimeout != time.Minute*15 {
		t.Fatalf("session manager idle timeout is incorrect: expectec %s got %s", (time.Minute * 15), (manager.IdleTimeout))
	}
	if manager.Lifetime != time.Hour*2 {
		t.Fatalf("session manager idle timeout is incorrect: expectec %s got %s", (time.Hour * 2), (manager.Lifetime))
	}
	if manager.RenewalTimeout != time.Minute {
		t.Fatalf("session manager idle timeout is incorrect: expectec %s got %s",
			(time.Minute), (manager.RenewalTimeout))
	}
	if reflect.TypeOf(manager.Store) == nil {
		t.Fatalf("session manager store is not initialized")
	}
}

// TestSessionManager_WatchTimeouts tests the session timeout handling, including renewal and invalidation.

// renewalTime, idleTime and lifetime are not up yet -- only the idleTime should be reset
func TestNoTimeExpired(t *testing.T) {

	manager.Store = NewMockStorage()

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 5

	session := manager.CreateSession()
	idleTime := session.IdleTime

	session = manager.WatchTimeouts(session)

	if session.IdleTime == idleTime {
		t.Fatalf("Session idle time was not reset;\n Expecting: \t %v \n Got: \t\t %v",
			session.IdleTime.Add(manager.IdleTimeout), session.IdleTime)
	}

}

// renewalTime is up but idleTime and lifeTime not up yet -- reset sessionID, idleTime and renewalTime
func TestRenewalTimeExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 5

	session := manager.CreateSession()

	renewalTime, idleTime, lifeTime := session.RenewalTime, session.IdleTime, session.LifeTime
	sessionID := session.ID
	time.Sleep(manager.RenewalTimeout)

	session = manager.WatchTimeouts(session)

	if time.Now().Before(idleTime) && time.Now().Before(lifeTime) && time.Now().After(renewalTime) {
		if session.RenewalTime.After(renewalTime) && session.IdleTime.After(idleTime) {

		} else {
			t.Fatalf("Either or both session renewal time or idle time was not reset\n Got:\n\t renewal Time: %v \t		 idleTim:: %v Expected: \n\t renewalTime: %v idleTime: %v\n",
				session.RenewalTime, session.IdleTime,
				renewalTime.Add(manager.RenewalTimeout), idleTime.Add(manager.IdleTimeout))
		}
	} else {
		t.Fatalf("idle time has elapsed but so has lifetime and renwalTime: cannot run test")
	}

	if session.ID == sessionID {
		t.Fatal("sessionID was not reset")
	}

}

// renewalTime and idleTime are up but lifeTime is not up yet -- invalidate session
// i.e. cookie value == "", cookie.expiryTime is in the past,  session was deleted from store,
func TestRenewalAndIdleTimeExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 5

	session := manager.CreateSession()

	// renewalTime, idleTime, lifeTime := session.RenewalTime, session.IdleTime, session.LifeTime
	sessionID := session.ID
	cookieExpiryTime := session.Cookie.Expires
	time.Sleep(manager.IdleTimeout)

	session = manager.WatchTimeouts(session)

	if time.Now().After(session.LifeTime) {
		t.Error("session lifetime has also expired; it should not have")
	}
	if session.Cookie.Value != "" {
		t.Fatalf("session cookie value should be empty \n expected: \n got: \t %v", session.Cookie.Value)
	}
	if time.Now().Before(session.Cookie.Expires) {
		t.Fatalf("session cookie should have expired \n expected: %v \n got: \t %v", cookieExpiryTime, session.Cookie.Expires)
	}
	if _, err := manager.Store.Get(sessionID); err == nil {
		t.Fatalf("session was not deleted from the store")
	}

}

// renewalTime, idleTime and lifetime are all up -- invalidate session
func TestAllTimesExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 1
	manager.Lifetime = time.Second * 2

	session := manager.CreateSession()

	sessionID := session.ID
	cookieExpiryTime := session.Cookie.Expires
	time.Sleep(manager.Lifetime)

	session = manager.WatchTimeouts(session)

	if time.Now().Before(session.IdleTime) && time.Now().Before(session.LifeTime) && time.Now().After(session.RenewalTime) {
		t.Error("some or all session times have not expired yet")
	}
	if session.Cookie.Value != "" {
		t.Fatalf("session cookie value should be empty \n expected: \n got: \t %v", session.Cookie.Value)
	}
	if time.Now().Before(session.Cookie.Expires) {
		t.Fatalf("session cookie should have expired \n expected: %v \n got: \t %v", cookieExpiryTime, session.Cookie.Expires)
	}
	if _, err := manager.Store.Get(sessionID); err == nil {
		t.Fatalf("session was not deleted from the store")
	}

}

// renewalTime and lifetime are all up but idleTime is not -- invalidate session
func TestRenewalAndLifeTimeExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 3

	session := manager.CreateSession()

	session = manager.WatchTimeouts(session)

	sessionID := session.ID
	cookieExpiryTime := session.Cookie.Expires

	session = manager.WatchTimeouts(session)

	time.Sleep(manager.IdleTimeout)
	session.RenewalTime = time.Now().Add(manager.RenewalTimeout)
	session = manager.WatchTimeouts(session)

	if time.Now().After(session.IdleTime) && time.Now().Before(session.LifeTime) && time.Now().After(session.RenewalTime) {
		t.Error("renewal time, and or lifetime has not expired yet")
	}
	if session.Cookie.Value != "" {
		t.Fatalf("session cookie value should be empty \n expected: \n got: \t %v", session.Cookie.Value)
	}
	if time.Now().Before(session.Cookie.Expires) {
		t.Fatalf("session cookie should have expired \n expected: %v \n got: \t %v", cookieExpiryTime, session.Cookie.Expires)
	}
	if _, err := manager.Store.Get(sessionID); err == nil {
		t.Fatalf("session was not deleted from the store")
	}

}

// lifetime is up but renewalTime and idleTime is not -- invalidate session
func TestLifetimeExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 1
	manager.Lifetime = time.Second * 2

	session := manager.CreateSession()

	session = manager.WatchTimeouts(session)

	sessionID := session.ID
	cookieExpiryTime := session.Cookie.Expires

	// by this time the lifetime should be expired
	// while the idletime and renewal time should have been reset
	session = manager.WatchTimeouts(session)
	time.Sleep(manager.RenewalTimeout)
	session = manager.WatchTimeouts(session)
	time.Sleep(manager.RenewalTimeout)
	session = manager.WatchTimeouts(session)

	if time.Now().After(session.LifeTime) && time.Now().Before(session.IdleTime) && time.Now().After(session.RenewalTime) {
		t.Fatal("renewal time, and or idle time has not expired yet")
	}
	if session.Cookie.Value != "" {
		t.Fatalf("session cookie value should be empty \n expected: \n got: \t %v", session.Cookie.Value)
	}
	if time.Now().Before(session.Cookie.Expires) {
		t.Fatalf("session cookie should have expired \n expected: %v \n got: \t %v", cookieExpiryTime, session.Cookie.Expires)
	}
	if _, err := manager.Store.Get(sessionID); err == nil {
		t.Fatalf("session was not deleted from the store")
	}
}

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
	cookieExpiryTime := session.Cookie.Expires

	session = manager.InvalidateSession(session)

	if session.Cookie.Value != "" {
		t.Fatalf("session cookie value should be empty \n expected: \n got: \t %v", session.Cookie.Value)
	}
	if time.Now().Before(session.Cookie.Expires) {
		t.Fatalf("session cookie should have expired \n expected: %v \n got: \t %v", cookieExpiryTime, session.Cookie.Expires)
	}
	if _, err := manager.Store.Get(sessionID); err == nil {
		t.Fatalf("session was not deleted from the store")
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
	cookie := session.Cookie

	session = manager.RenewSession(session)

	if sessionID == session.ID {
		t.Fatal("session id was not renewed")
	}
	if cookie.Value == session.Cookie.Value {
		t.Fatal("cookie value was not renewed")
	}
	if _, err := manager.Store.Get(sessionID); err == nil {
		t.Fatal("old session id not updated in the store")
	}
	if _, err := manager.Store.Get(session.ID); err != nil {
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
