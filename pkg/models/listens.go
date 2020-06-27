package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// ListenModel is our interface with the listens table.
type ListenModel struct {
	DB *gorm.DB
}

// Create adds a row in the episodes table.
func (m *ListenModel) Create(userID, episodeID uint) error {
	var blank Listen
	listen := &Listen{
		UserID:     userID,
		EpisodeID:  episodeID,
		ListenedAt: time.Now(),
	}

	err := m.DB.Model(&listen).Where("user_id = ? AND episode_id = ?", userID, episodeID).First(&blank).Error

	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err = m.DB.Create(&listen).Error
		}
	} else {
		err = m.DB.Model(&listen).Where("user_id = ? AND episode_id = ?", userID, episodeID).Updates(&listen).Error
	}

	if err != nil {
		return err
	}

	return nil
}

// FindAll gets all listens for the given user ID.
func (m *ListenModel) FindAll(userID uint) ([]Listen, error) {
	var listens []Listen

	err := m.DB.Where("user_id = ?", userID).Find(&listens).Error
	if err != nil {
		return listens, err
	}

	return listens, nil
}

// FindByPodcast gets all listens for the given user ID and the given podcast ID.
func (m *ListenModel) FindByPodcast(userID uint, episodeIDs []uint) ([]Listen, error) {
	var listens []Listen

	err := m.DB.Where("user_id = ? AND episode_id IN (?)", userID, episodeIDs).Find(&listens).Error
	if err != nil {
		return listens, err
	}

	return listens, nil
}

// Delete removes a listen from the database.
func (m *ListenModel) Delete(userID, episodeID uint) error {
	var listen Listen

	return m.DB.Model(&listen).Where("user_id = ? AND episode_id = ?", userID, episodeID).Delete(Listen{}).Error
}
