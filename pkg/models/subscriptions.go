package models

import "github.com/jinzhu/gorm"

// SubscriptionModel represents our interface with the subscriptions table.
type SubscriptionModel struct {
	DB *gorm.DB
}

// Create inserts a new subscription into the database.
func (m *SubscriptionModel) Create(podcastID int, userID uint) error {
	var subscription Subscription

	// ... and save it to the database.
	err := m.DB.FirstOrCreate(&subscription, Subscription{
		PodcastID: podcastID,
		UserID:    userID,
	}).Error

	if err != nil {
		return m.DB.Error
	}

	return nil
}

// Find finds a subscription by collectionID and userID.
func (m *SubscriptionModel) Find(collectionID int, userID uint) (Subscription, error) {
	var subscription Subscription
	err := m.DB.First(&subscription, "podcast_id = ? AND user_id = ?", collectionID, userID).Error

	return subscription, err
}

// FindAll returns all subscriptions for a given userID.
func (m *SubscriptionModel) FindAll(userID uint) ([]Subscription, error) {
	var subscriptions, blank []Subscription

	err := m.DB.Preload("Podcast").Preload("Podcast.Episodes").Find(&subscriptions, "user_id = ?", userID).Error
	if err != nil {
		return blank, err
	}

	return subscriptions, nil
}

// Delete removes the provided subscription from the provided user.
func (m *SubscriptionModel) Delete(collectionID int, userID uint) error {
	return m.DB.Unscoped().Delete(Subscription{}, "podcast_id = ? AND user_id = ?", collectionID, userID).Error
}

// Podcast returns the podcast for the given subscription ID.
func (m *SubscriptionModel) Podcast(collectionID int) (Podcast, error) {
	var podcast Podcast
	err := m.DB.First(&podcast, "id = ?", collectionID).Error

	return podcast, err

}
