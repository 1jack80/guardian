package guardian

/*
	ERROR TYPES:
		- NOT FOUND
*/

type Storer interface {
	// retrieve the sesison data from the underlying container
	// and decode it before returning it to the calling function
	Get(sessionID string) (Session, error)

	// save an encoded form of the given session data into
	// the underlying container
	Save(session Session) error

	// delete the session identified by the given sessionID
	Delete(sessionID string) error

	// update parts of the session that identifes with the given sessionID:
	// the new session is used to replace the old session hance,
	// using this function requires that a pointer to the updated
	// copy of the old session is created an passed to this function.
	Update(sessionID string, newSession Session) error
}
