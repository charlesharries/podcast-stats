package models

import (
	"errors"
	"time"

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

// Podcast is a single podcast from iTunes.
type Podcast struct {
	ID       int `gorm:"primary_key"`
	Name     string
	Feed     string
	Episodes []Episode
}

// Subscription represents a relationship between a user and a podcast.
type Subscription struct {
	UserID    uint `gorm:"index:subscription_user_id"`
	PodcastID int  `gorm:"index:subscription_podcast_id"`
	Podcast   Podcast
}

// Episode is a single podcast episode.
type Episode struct {
	ID          uint   `gorm:"primary_key"`
	PodcastID   int    `gorm:"index:episode_podcast_id"`
	GUID        string `gorm:"type:varchar(100);unique_index"`
	Title       string
	Source      string
	PublishedOn time.Time
	Duration    int
}
