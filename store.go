package guardian

type container interface {

	// retrieve some data from the store
	// initialized when the store is created
	get(sessionID string) ([]byte, error)

	// save some session data to the store in a key value manner
	// returns nil if session was saved successfully
	put(sessionID string, data []byte) error

	// delete some session data from the store usign the sessionID
	// as an identifier
	delete(sessionID string) error
}

/*
	ERROR TYPES:
		- NOT FOUND
*/

type Store struct {
	codec   Codec
	storage container
}
type Codec interface {
	encode(data *Session) ([]byte, error)
	decode(data []byte) (*Session, error)
}

// retrieve the sesison data from the underlying container
// and decode it before returning it to the calling function
func (s *Store) Get(sessionID string) (*Session, error) {
	encodedData, err := s.storage.get(sessionID)
	if err != nil {
		return &Session{}, err
	}

	session, err := s.codec.decode(encodedData)
	if err != nil {
		return &Session{}, err
	}

	return session, nil
}

// save an encoded form of the given session data into
// the underlying container
func (s *Store) Save(session *Session) error {
	binaryData, err := s.codec.encode(session)
	if err != nil {
		return err
	}
	err = s.storage.put(session.id, binaryData)
	return err
}

// delete the session identified by the given sessionID
func (s *Store) Delete(sessionID string) error {
	return s.storage.delete(sessionID)
}

// update parts of the session that identifes with the given sessionID:
// the new session is used to replace the old session hance,
// using this function requires that a pointer to the updated
// copy of the old session is created an passed to this function.
func (s *Store) Update(sessionID string, newSession *Session) error {
	// TODO: find a better way to update sessions especially considering the case of renewing sessionIDs

	if err := s.storage.delete(sessionID); err != nil {
		return err
	}

	return s.Save(newSession)
}
