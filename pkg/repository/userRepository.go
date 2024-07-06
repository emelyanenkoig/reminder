package repository

import (
	"emelyanenkoig/reminder/pkg/cache"
	"emelyanenkoig/reminder/pkg/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	DB    *gorm.DB
	Cache *cache.Cache
}

func NewUserRepository(db *gorm.DB, cache *cache.Cache) *UserRepository {
	return &UserRepository{
		DB:    db,
		Cache: cache,
	}
}

func (repo *UserRepository) CreateUser(user *models.User) error {
	// create user in db
	err := repo.DB.Create(user).Error
	if err != nil {
		return err
	}

	repo.Cache.AddUser(user)
	return nil

}

func (repo *UserRepository) GetUserById(userID uint) (*models.User, error) {
	// lookup cache
	user, exist := repo.Cache.GetUser(userID)
	if exist {
		return user, nil
	}

	// lookup db
	user = &models.User{}
	err := repo.DB.Preload("Reminders").First(user, userID).Error
	if err != nil {
		return nil, err
	}

	// add in cache
	repo.Cache.AddUser(user)
	return user, nil
}
