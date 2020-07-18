package mysqlcache

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

// Model is how we interact with the cache
type Model struct {
	Expiry time.Duration
	DB     *gorm.DB
}

// CacheEntry is a single entry in the cache
type CacheEntry struct {
	ID      string `gorm:"primary_key;type:varchar(255);unique_index"`
	Val     string `gorm:"type:TEXT"`
	Created time.Time
}

// ErrCacheExpired is returned when a fetched entry has expired.
var ErrCacheExpired = errors.New("cache expired")

// ErrCacheMiss is returned when an entry isn't found in the cache.
var ErrCacheMiss = errors.New("cache miss")

// Set sets a row by key in the Redis cache.
func (m *Model) Set(id, val string) error {
	ce := &CacheEntry{
		ID:      id,
		Val:     val,
		Created: time.Now(),
	}

	err := m.DB.Where(CacheEntry{ID: ce.ID}).Assign(&ce).FirstOrCreate(&ce).Error
	if err != nil {
		return m.DB.Error
	}

	return nil
}

// Get returns the value at a key, along with the time
// that that key was set.
func (m *Model) Get(id string) (string, error) {
	var ce CacheEntry

	err := m.DB.Where("id = ?", id).First(&ce).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrCacheMiss
	}

	if ce.Created.Add(m.Expiry).Before(time.Now()) {
		// TODO(charles): Remove the item from the cache

		return "", ErrCacheExpired
	}

	return ce.Val, nil
}
