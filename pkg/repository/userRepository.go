package repository

import (
	"emelyanenkoig/reminder/pkg/models"
	"errors"
	"sync"
)

type UserRepository interface {
	GetUserByID(userID uint) (*models.User, error)
	CreateUser(user *models.User) error
}

type userRepository struct {
	users     map[uint]*models.User
	usersLock sync.RWMutex
}

func NewUserRepository() UserRepository {
	return &userRepository{
		users: make(map[uint]*models.User),
	}
}

func (r *userRepository) GetUserByID(userID uint) (*models.User, error) {
	r.usersLock.RLock()
	defer r.usersLock.RUnlock()

	user, ok := r.users[userID]
	if !ok {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (r *userRepository) CreateUser(user *models.User) error {
	r.usersLock.Lock()
	defer r.usersLock.Unlock()

	// Check if user already exists
	if _, ok := r.users[user.ID]; ok {
		return errors.New("user already exists")
	}

	// Add user to repository
	r.users[user.ID] = user

	return nil
}
