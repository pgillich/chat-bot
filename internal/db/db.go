// Package db provides DB
package db

import (
	"fmt"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres" // import Postgres driver
	log "github.com/sirupsen/logrus"

	"github.com/pgillich/chat-bot/internal/logger"
)

// User table
type User struct {
	gorm.Model
	UID    string `gorm:"uniq_key"`
	Name   string
	BornOn *time.Time
	BornAt string
}

// TableName forces table name singular
func (User) TableName() string {
	return "user"
}

// DbHandler is an interface for DB backend (and faking)
type DbHandler interface { // nolint:golint
	Connect() error
	Close()
	GetOrCreateUser(uid string) (User, error)
	Update(user User) error
}

// RealDbHandler is a real implementation of DbHandler
type RealDbHandler struct {
	Host     string
	User     string
	Database string
	Password string

	db *gorm.DB

	mx *sync.RWMutex
}

// Connect connects to the DB
func (dbHandler *RealDbHandler) Connect() error {
	var err error

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s sslmode=disable password=%s",
		dbHandler.Host, dbHandler.User, dbHandler.Database, dbHandler.Password)
	if dbHandler.db, err = gorm.Open("postgres", dbURI); err != nil {
		return err
	}

	if logger.Get().IsLevelEnabled(log.DebugLevel) {
		dbHandler.db = dbHandler.db.Debug()
	}

	dbHandler.db = dbHandler.db.AutoMigrate(&User{})
	if dbHandler.db.Error != nil {
		return dbHandler.db.Error
	}

	dbHandler.mx = new(sync.RWMutex)

	return nil
}

// Close closes DB connection
func (dbHandler *RealDbHandler) Close() {
	dbHandler.mx.Lock()
	defer dbHandler.mx.Unlock()

	if err := dbHandler.db.Close(); err != nil {
		logger.Get().Warning("database conn close error, ", err)
	}
}

// GetOrCreateUser creates a new user or returns, if exists
func (dbHandler *RealDbHandler) GetOrCreateUser(uid string) (User, error) {
	dbHandler.mx.RLock()
	defer dbHandler.mx.RUnlock()

	templateUser := User{UID: uid}
	user := User{}

	db := dbHandler.db.Where(templateUser).FirstOrCreate(&user)
	if db.Error != nil {
		return templateUser, db.Error
	}

	return user, nil
}

// Update updates user
func (dbHandler *RealDbHandler) Update(user User) error { // nolint:gocritic
	dbHandler.mx.RLock()
	defer dbHandler.mx.RUnlock()

	db := dbHandler.db.Save(user)
	if db.Error != nil {
		return db.Error
	}

	return nil
}
