package repository

import (
	"emelyanenkoig/reminder/pkg/cache"
	"emelyanenkoig/reminder/pkg/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserById(userID uint) (*models.User, error)
	GetUsers() ([]models.User, error)
	UpdateUser(userID uint, updatedUser *models.User) error
	DeleteUser(userID uint) error
}

type userRepository struct {
	DB    *gorm.DB
	Cache *cache.Cache
}

func NewUserRepository(db *gorm.DB, cache *cache.Cache) *userRepository {
	return &userRepository{
		DB:    db,
		Cache: cache,
	}
}

func (repo *userRepository) CreateUser(user *models.User) error {
	if err := repo.DB.Create(user).Error; err != nil {
		return err
	}
	go repo.Cache.AddUser(user)
	return nil
}

func (repo *userRepository) GetUserById(userID uint) (*models.User, error) {
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

func (repo *userRepository) GetUsers() ([]models.User, error) {
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

func (repo *userRepository) UpdateUser(userID uint, updatedUser *models.User) error {
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

func (repo *userRepository) DeleteUser(userID uint) error {
	// Удаление связанных напоминаний
	if err := repo.DB.Where("user_id = ?", userID).Delete(&models.Reminder{}).Error; err != nil {
		return err
	}

	user := models.User{ID: userID}
	if err := repo.DB.Delete(&user).Error; err != nil {
		return err
	}

	go repo.Cache.DeleteUser(userID)

	return nil
}
