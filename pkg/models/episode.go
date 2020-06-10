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
func (m *EpisodeModel) Create(title, guid string, podcastID int, publishedOn time.Time, duration float64) error {
	episode := &Episode{
		Title:       title,
		GUID:        guid,
		PodcastID:   podcastID,
		PublishedOn: publishedOn,
		Duration:    duration,
	}

	err := m.DB.Where(Episode{GUID: guid}).Assign(&episode).FirstOrCreate(&episode).Error
	if err != nil {
		return err
	}

	return nil
}
