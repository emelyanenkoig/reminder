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
	if err := repo.DB.Create(user).Error; err != nil {
		return err
	}
	repo.Cache.AddUser(user)
	return nil
}

func (repo *UserRepository) GetUserById(userID uint) (*models.User, error) {
	if user, exist := repo.Cache.GetUser(userID); exist {
		return user, nil
	}

	user := &models.User{}
	if err := repo.DB.Preload("Reminders").First(user, userID).Error; err != nil {
		return nil, err
	}
	repo.Cache.AddUser(user)
	return user, nil
}
