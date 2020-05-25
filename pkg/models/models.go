package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// ErrNoRecord displays if a user requests a resource that doesn't exist
var ErrNoRecord = errors.New("models: no matching record found")

// ErrInvalidCredentials will be displayed if a user tries to access a
// resource that they aren't allowed to.
var ErrInvalidCredentials = errors.New("models: invalid credentials")

// ErrDuplicateEmail will be displayed if a user tries to signup with an
// email that has already been taken.
var ErrDuplicateEmail = errors.New("models: duplicate email")

// User represents the schema for our user in the database.
type User struct {
	gorm.Model
	Email         string `gorm:"type:varchar(100);unique_index;not null"`
	Password      []byte `gorm:"type:varchar(60);not null"`
	Subscriptions []Subscription
}

// Subscription represents a relationship between a user and a podcast.
type Subscription struct {
	gorm.Model
	UserID       uint `gorm:"index:subscription_user_id"`
	CollectionID int  `gorm:"index:subscription_collection_id"`
}
