package repository

import "github.com/drama-generator/backend/domain/models"

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
	FindByID(id uint) (*models.User, error)
	Create(user *models.User) error
	UpdatePassword(userID uint, hash string) error
	CreateWithInitialCredits(user *models.User, initialCredits int) error
}
