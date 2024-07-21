package services

import (
	"emelyanenkoig/reminder/pkg/models"
	"emelyanenkoig/reminder/pkg/repository"
)

type ReminderService struct {
	repo repository.ReminderRepository
}

func NewReminderService(repo repository.ReminderRepository) *ReminderService {
	return &ReminderService{repo: repo}
}

func (s *ReminderService) CreateReminder(reminder *models.Reminder) error {
	return s.repo.CreateReminder(reminder)
}

func (s *ReminderService) GetReminderByID(reminderId, id uint) (*models.Reminder, error) {
	return s.repo.GetReminderByUserId(id, reminderId)
}

func (s *ReminderService) GetRemindersByUserID(userID uint) ([]models.Reminder, error) {
	return s.repo.GetRemindersByUser(userID)
}

func (s *ReminderService) UpdateReminder(userID, reminderID uint, reminder *models.Reminder) error {
	return s.repo.UpdateReminder(userID, reminderID, reminder)
}

func (s *ReminderService) DeleteReminder(userId, reminderID uint) error {
	return s.repo.DeleteReminder(userId, reminderID)
}
