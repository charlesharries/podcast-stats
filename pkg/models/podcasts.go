package models

import (
	"github.com/jinzhu/gorm"
)

// PodcastModel is our way to interact with the podcasts table in the DB.
type PodcastModel struct {
	DB *gorm.DB
}

// Create inserts a new podcast into the database.
func (m *PodcastModel) Create(ID int, collectionName, feed string) error {
	podcast := &Podcast{
		ID:   ID,
		Name: collectionName,
		Feed: feed,
	}

	err := m.DB.Where(Podcast{ID: ID}).Assign(&podcast).FirstOrCreate(&podcast).Error
	if err != nil {
		return m.DB.Error
	}

	return nil
}

// Find gets a single podcast by collectionID.
func (m *PodcastModel) Find(collectionID int) (Podcast, error) {
	var podcast Podcast
	err := m.DB.Preload("Episodes").First(&podcast, "id = ?", collectionID).Error

	return podcast, err
}
