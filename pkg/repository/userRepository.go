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
	go repo.Cache.AddUser(user)
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
	go repo.Cache.AddUser(user)
	return user, nil
}

func (repo *UserRepository) GetUsers() ([]models.User, error) {
	users := repo.Cache.GetUsers()
	if len(users) > 0 {
		return users, nil
	}

	var dbUsers []models.User
	if err := repo.DB.Preload("Reminders").Find(&dbUsers).Error; err != nil {
		return nil, err
	}

	go func() {
		for _, user := range dbUsers {
			repo.Cache.AddUser(&user)
		}
	}()

	return dbUsers, nil
}

func (repo *UserRepository) UpdateUser(userID uint, updatedUser *models.User) error {
	// Update user in the DB
	err := repo.DB.Save(updatedUser).Error
	if err != nil {
		return err
	}

	// Update reminders (if any)
	for _, reminder := range updatedUser.Reminders {
		err := repo.DB.Save(&reminder).Error
		if err != nil {
			return err
		}
	}

	// Update cache
	go func() {
		repo.Cache.DeleteUser(userID)
		repo.Cache.AddUser(updatedUser)
	}()
	return nil
}

func (repo *UserRepository) DeleteUser(userID uint) error {
	user := models.User{ID: userID}

	if err := repo.DB.Delete(&user).Error; err != nil {
		return err
	}

	go repo.Cache.DeleteUser(userID)

	return nil
}
