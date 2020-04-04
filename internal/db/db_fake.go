package db

import (
	"sync"
)

// FakeDbHandler is a fake implementation of DbHandler
type FakeDbHandler struct {
	data sync.Map
}

// Connect resets everything
func (dbHandler *FakeDbHandler) Connect() error {
	dbHandler.data = sync.Map{}

	return nil
}

// Close deletes DB
func (dbHandler *FakeDbHandler) Close() {
	dbHandler.data.Range(func(key interface{}, value interface{}) bool {
		dbHandler.data.Delete(key)

		return true
	})
}

// GetOrCreateUser creates a new user or returns, if exists
func (dbHandler *FakeDbHandler) GetOrCreateUser(uid string) (User, error) {
	userIf, _ := dbHandler.data.LoadOrStore(uid, User{UID: uid})
	user, _ := userIf.(User) // nolint:errcheck

	return user, nil
}

// Update updates user
func (dbHandler *FakeDbHandler) Update(user User) error { // nolint:gocritic
	dbHandler.data.Store(user.UID, user)

	return nil
}
