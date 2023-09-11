package guardian_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/1jack80/guardian"
)

var (
	manager, manager_err = guardian.New("test_manager")
)

// TestSessionManager_CreateSession tests the creation of a new session and validates its attributes.
func TestSessionManager_CreateSession(t *testing.T) {
	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	if manager.ContextKey == "" {
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

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 5

	session := manager.CreateSession()
	idleTime := session.IdleTime

	manager.WatchTimeouts(&session)

	if session.IdleTime == idleTime {
		t.Fatalf("Session idle time was not reset;\n Expecting: \t %v \n Got: \t\t %v",
			session.IdleTime.Add(manager.IdleTimeout), session.IdleTime)
	}

}

/*
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

	manager.WatchTimeouts(&session)

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

	manager.WatchTimeouts(&session)

}

// renewalTime, idleTime and lifetime are all up -- invalidate session
func TestAllTimesExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 5

	session := manager.CreateSession()

	manager.WatchTimeouts(&session)

}

// renewalTime and lifetime are all up but idleTime is not -- invalidate session
func TestRenewalAndLifeTimeExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 5

	session := manager.CreateSession()

	manager.WatchTimeouts(&session)

}

// lifetime is up but renewalTime and idleTime is not -- invalidate session
func TestLifetimeExpired(t *testing.T) {

	if manager_err != nil {
		t.Fatalf("unable to create session manager: err -- %s", manager_err.Error())
	}

	manager.RenewalTimeout = time.Second * 1
	manager.IdleTimeout = time.Second * 2
	manager.Lifetime = time.Second * 5

	session := manager.CreateSession()

	manager.WatchTimeouts(&session)

}

// TestSessionManager_PopulateRequestContext ensures that session data is correctly populated in the request context.
func TestSessionManager_PopulateRequestContext(t *testing.T) {
	// Test logic here
}

// TestSessionManager_InvalidateSession tests the invalidation of a session and verifies that the session data is removed from the store and the session cookie is expired.
func TestSessionManager_InvalidateSession(t *testing.T) {
	// Test logic here
}

// TestStore_Get tests the retrieval of session data from the store.
func TestStore_Get(t *testing.T) {
	// Test logic here
}

// TestStore_Save tests saving session data to the store and verifies that it's stored correctly.
func TestStore_Save(t *testing.T) {
	// Test logic here
}

// TestStore_Delete tests deleting session data from the store.
func TestStore_Delete(t *testing.T) {
	// Test logic here
}

// TestStore_Update tests updating session data in the store.
func TestStore_Update(t *testing.T) {
	// Test logic here
}

// Tests for Objectives Not Yet Implemented:

// TestSessionManager_RenewSession tests the session renewal functionality once implemented.
func TestSessionManager_RenewSession(t *testing.T) {
	// Test logic here
}

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

// TestSessionManager_Documentation is not a code test but ensures that your documentation (e.g., Go doc comments) is accurate and comprehensive.
func TestSessionManager_Documentation(t *testing.T) {
	// Test logic here
}

// Integration Tests: Consider integration tests where you simulate multiple sessions and interactions between them to ensure the session manager behaves correctly in a real-world scenario.

// Edge Cases: Test edge cases such as when session creation or timeout handling is triggered under various conditions (e.g., minimum timeout values, maximum session data size).

// TestNamespaceManager: If you have not yet implemented multiple session namespaces, create tests to ensure that multiple namespaces can coexist and function correctly.


*/
