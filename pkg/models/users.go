package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// UserModel represents our connection to the `users` table
// in the database. It's the model.
type UserModel struct {
	DB *gorm.DB
}

// Create inserts a new user into the database.
func (m *UserModel) Create(email, password string) error {
	// Hash the user's password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	// Mock up the user...
	user := &User{
		Email:    email,
		Password: hashedPassword,
	}

	// ...and save them to the database.
	err = m.DB.Create(user).Error
	if err != nil {
		return m.DB.Error
	}

	return nil
}

// Authenticate finds a user in the database with the email and
// password provided.
func (m *UserModel) Authenticate(email, password string) (User, error) {
	var user, blank User

	err := m.DB.First(&user, "email = ?", email).Error
	if err != nil {
		return blank, err
	}

	err = bcrypt.CompareHashAndPassword(user.Password, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return blank, ErrInvalidCredentials
		}

		return blank, err
	}

	return user, nil
}
