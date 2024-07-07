package repository

import (
	"emelyanenkoig/reminder/pkg/cache"
	"emelyanenkoig/reminder/pkg/models"
	"gorm.io/gorm"
	"time"
)

type ReminderRepository struct {
	DB    *gorm.DB
	Cache *cache.Cache
}

func NewReminderRepository(db *gorm.DB, cache *cache.Cache) *ReminderRepository {
	return &ReminderRepository{
		DB:    db,
		Cache: cache,
	}
}

func (repo *ReminderRepository) CreateReminder(reminder *models.Reminder) error {
	userId := reminder.UserID
	err := repo.DB.Create(&reminder).Error
	if err != nil {
		return err
	}

	go repo.Cache.AddReminder(userId, reminder)
	return nil
}

func (repo *ReminderRepository) GetReminderByUserId(userId uint, reminderId uint) (*models.Reminder, error) {
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

func (repo *ReminderRepository) GetRemindersByUser(userId uint) ([]models.Reminder, error) {
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

func (repo *ReminderRepository) UpdateReminder(userID uint, reminderID uint, updatedReminder *models.Reminder) error {
	_, found := repo.Cache.GetReminderByUserId(userID, reminderID)
	if !found {
		err := repo.DB.Where("user_id = ? AND id = ?", userID, reminderID).First(updatedReminder).Error
		if err != nil {
			return err
		}
	}

	updatedReminder.UserID = userID
	updatedReminder.ID = reminderID
	updatedReminder.DueDate = time.Now()

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

func (repo *ReminderRepository) DeleteReminder(userID uint, reminderID uint) error {
	err := repo.DB.Where("user_id = ? AND id = ?", userID, reminderID).Delete(&models.Reminder{}).Error
	if err != nil {
		return err
	}
	go repo.Cache.DeleteReminder(userID, reminderID)
	return nil
}
