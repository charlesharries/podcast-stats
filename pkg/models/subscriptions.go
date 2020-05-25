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
