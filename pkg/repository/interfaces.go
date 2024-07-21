package repository

import "emelyanenkoig/reminder/pkg/models"

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserById(userID uint) (*models.User, error)
	GetUsers() ([]models.User, error)
	UpdateUser(userID uint, updatedUser *models.User) error
	DeleteUser(userID uint) error
}

type ReminderRepository interface {
	CreateReminder(reminder *models.Reminder) error
	GetReminderByUserId(userId uint, reminderId uint) (*models.Reminder, error)
	GetRemindersByUser(userId uint) ([]models.Reminder, error)
	UpdateReminder(userID uint, reminderID uint, updatedReminder *models.Reminder) error
	DeleteReminder(userID uint, reminderID uint) error
}
