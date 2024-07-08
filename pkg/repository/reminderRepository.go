package repository

import (
	"emelyanenkoig/reminder/pkg/cache"
	"emelyanenkoig/reminder/pkg/models"
	"errors"
	"gorm.io/gorm"
	"time"
)

type ReminderRepository interface {
	CreateReminder(reminder *models.Reminder) error
	GetReminderByUserId(userId uint, reminderId uint) (*models.Reminder, error)
	GetRemindersByUser(userId uint) ([]models.Reminder, error)
	UpdateReminder(userID uint, reminderID uint, updatedReminder *models.Reminder) error
	DeleteReminder(userID uint, reminderID uint) error
}

type reminderRepository struct {
	DB    *gorm.DB
	Cache *cache.Cache
}

func NewReminderRepository(db *gorm.DB, cache *cache.Cache) *reminderRepository {
	return &reminderRepository{
		DB:    db,
		Cache: cache,
	}
}

// Валидация данных напоминания
func validateReminder(reminder *models.Reminder) error {
	if reminder.Title == "" {
		return errors.New("title is required")
	}
	if reminder.DueDate.Before(time.Now()) {
		return errors.New("due date cannot be in the past")
	}
	return nil
}

func (repo *reminderRepository) CreateReminder(reminder *models.Reminder) error {
	if err := validateReminder(reminder); err != nil {
		return err
	}
	userId := reminder.UserID
	err := repo.DB.Create(&reminder).Error
	if err != nil {
		return err
	}

	go repo.Cache.AddReminder(userId, reminder)
	return nil
}

func (repo *reminderRepository) GetReminderByUserId(userId uint, reminderId uint) (*models.Reminder, error) {
	reminder, found := repo.Cache.GetReminderByUserId(userId, reminderId)
	if found {
		return reminder, nil
	}

	err := repo.DB.Where("user_id = ? AND id = ?", userId, reminderId).First(&reminder).Error
	if err != nil {
		return nil, err
	}
	return reminder, nil
}

func (repo *reminderRepository) GetRemindersByUser(userId uint) ([]models.Reminder, error) {
	reminders, found := repo.Cache.GetRemindersListByUser(userId)
	if found {
		return reminders, nil
	}

	err := repo.DB.Where("user_id = ?", userId).Find(&reminders).Error
	if err != nil {
		return nil, err
	}
	return reminders, nil
}

func (repo *reminderRepository) UpdateReminder(userID uint, reminderID uint, updatedReminder *models.Reminder) error {
	if err := validateReminder(updatedReminder); err != nil {
		return err
	}
	_, found := repo.Cache.GetReminderByUserId(userID, reminderID)
	if !found {
		err := repo.DB.Where("user_id = ? AND id = ?", userID, reminderID).First(updatedReminder).Error
		if err != nil {
			return err
		}
	}

	updatedReminder.UserID = userID
	updatedReminder.ID = reminderID
	updatedReminder.DueDate = time.Now().Add(time.Hour * 24)

	err := repo.DB.Model(&models.Reminder{}).
		Where("user_id = ? AND id = ?", userID, reminderID).
		Updates(models.Reminder{
			Title:       updatedReminder.Title,
			Description: updatedReminder.Description,
			DueDate:     updatedReminder.DueDate,
		}).Error
	if err != nil {
		return err
	}

	go func() {
		if found {
			repo.Cache.DeleteReminder(userID, reminderID)
		}
		repo.Cache.AddReminder(userID, updatedReminder)
	}()
	return nil
}

func (repo *reminderRepository) DeleteReminder(userID uint, reminderID uint) error {
	err := repo.DB.Where("user_id = ? AND id = ?", userID, reminderID).Delete(&models.Reminder{}).Error
	if err != nil {
		return err
	}
	repo.Cache.DeleteReminder(userID, reminderID)
	return err
}
