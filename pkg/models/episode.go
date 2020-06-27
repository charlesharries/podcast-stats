package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// EpisodeModel is our interface with the episodes table.
type EpisodeModel struct {
	DB *gorm.DB
}

// Create adds a row in the episodes table.
func (m *EpisodeModel) Create(title, guid, source string, duration, podcastID int, publishedOn time.Time) error {
	episode := &Episode{
		Title:       title,
		GUID:        guid,
		Source:      source,
		Duration:    duration,
		PodcastID:   podcastID,
		PublishedOn: publishedOn,
	}

	err := m.DB.Where(Episode{GUID: guid}).Assign(&episode).FirstOrCreate(&episode).Error
	if err != nil {
		return err
	}

	return nil
}
