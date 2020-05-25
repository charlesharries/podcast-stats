package models

import "github.com/jinzhu/gorm"

// SubscriptionModel represents our interface with the subscriptions table.
type SubscriptionModel struct {
	DB *gorm.DB
}

// Create inserts a new subscription into the database.
func (m *SubscriptionModel) Create(collectionID int, userID uint) error {
	// Mock up the subscription...
	subscription := &Subscription{
		CollectionID: collectionID,
		UserID:       userID,
	}

	// ... and save it to the database.
	err := m.DB.Create(subscription).Error
	if err != nil {
		return m.DB.Error
	}

	return nil
}

// Get finds a subscription by collectionID and userID.
func (m *SubscriptionModel) Get(collectionID int, userID uint) (Subscription, error) {
	var subscription, blank Subscription

	err := m.DB.First(&subscription, "collection_id= ? AND user_id = ?", collectionID, userID).Error
	if err != nil {
		return blank, err
	}

	return subscription, nil
}
