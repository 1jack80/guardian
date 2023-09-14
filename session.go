package guardian

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const (
	INVALID int = iota
	VALID
)

type Session struct {
	ID         string
	Data       map[string]interface{}
	Status     int
	IdleTime   time.Time
	ExpiryTime time.Time
}

type contextKey string

type Manager struct {
	name          string
	store         Storer
	contextKey    contextKey
	cookieName    string
	idleTimeout   time.Duration
	expiryTimeout time.Duration
}

func NewManager(name string, store Storer) (Manager, error) {
	err := ValidateNamespace(name)
	if err != nil {
		return Manager{}, err
	}
	return Manager{
		name:          name,
		store:         store,
		cookieName:    name + "_session",
		contextKey:    newContextKey(name),
		idleTimeout:   time.Minute * 3,
		expiryTimeout: time.Hour * 2,
	}, err
}

// create a new session and add it to the store.
func (man *Manager) CreateSession() Session {
	newSession := Session{
		ID:         man.newSessionID(),
		Data:       make(map[string]interface{}),
		Status:     VALID,
		IdleTime:   time.Now().Add(man.idleTimeout),
		ExpiryTime: time.Now().Add(man.expiryTimeout),
	}
	man.store.Save(newSession)
	return newSession
}

func (man *Manager) SaveSession(sessonInstance Session) error {
	return man.store.Save(sessonInstance)
}

func (man *Manager) GetSession(sessionID string) (Session, error) {
	return man.store.Get(sessionID)
}

func (man *Manager) UpdateSession(sessionID string, sessionInstance Session) error {
	return man.store.Update(sessionID, sessionInstance)
}

// change the session id of the session but maintain the data therein
func (man *Manager) RenewSession(sessionID string) (Session, error) {
	newID := man.newSessionID()
	oldSession, err := man.store.Get(sessionID)
	if err != nil {
		return Session{}, err
	}
	oldSession.ID = newID

	err = man.store.Save(oldSession)
	if err != nil {
		return Session{}, err
	}

	err = man.store.Delete(sessionID)
	if err != nil {
		return oldSession, errors.New("unable to delete old session from store; although new session ID was saved successfully")
	}
	return oldSession, nil
}

// mark the session as invalid but keep it around until the session expiry time elapses
// by this time the associated cookie should have also expired then the session can be deleted
func (man *Manager) InvalidateSession(sessionID string) error {
	session, err := man.store.Get(sessionID)
	if err != nil {
		return err
	}
	session.Status = INVALID
	session.IdleTime = time.Now().Add(-man.idleTimeout)
	return man.store.Update(sessionID, session)
}

// a wrapper over the delete method in the store
func (man *Manager) DeleteSession(sessionID string) error {
	return man.store.Delete(sessionID)
}

// creates and returns a new cookie using a session
func (man *Manager) CeateCookie(sessionID string) (http.Cookie, error) {
	session, err := man.store.Get(sessionID)
	if err != nil {
		return http.Cookie{}, err
	}
	return http.Cookie{
		Name:    man.cookieName,
		Value:   sessionID,
		Expires: session.ExpiryTime,
		// Secure:   true, // https traffic only
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

// fill the request context with the given session and returns the updated request
func (man *Manager) PopulateRequestContext(r *http.Request, session Session) *http.Request {
	ctx := context.WithValue(r.Context(), man.contextKey, session)
	return r.WithContext(ctx)
}

// populates the contexts of new requests with the sessions to which the request cookie
// points. The middleware also extends the session idle times after each request
func (man *Manager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(man.cookieName)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		sessionID := cookie.Value
		session, err := man.store.Get(sessionID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// watch timeouts
		if time.Now().After(session.ExpiryTime) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if time.Now().After(session.IdleTime) {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		r = man.PopulateRequestContext(r, session)

		next.ServeHTTP(w, r)

		// session must be refetched from store in case other handlers down the chain
		// tampered with the current one.
		session, err = man.store.Get(sessionID)
		if err != nil {
			// http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			// log err
			return
		}

		if session.Status == VALID {
			session.IdleTime = time.Now().Add(man.idleTimeout)
			man.store.Update(sessionID, session)
		}

		newCookie, err := man.CeateCookie(sessionID)
		if err != nil {
			// http.Error(w, "Unable to respond with proper cookie: "+
			// http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			// log err
			return
		}
		http.SetCookie(w, &newCookie)
	})
}

// acts as an accessor to get the manager's context key as it must not be changed
func (man *Manager) ContextKey() contextKey {
	return man.contextKey
}
