package repository

import (
	"emelyanenkoig/reminder/pkg/models"
	"errors"
	"log"
	"sync"
)

type ReminderRepository interface {
	GetUserReminders(userID uint) ([]models.Reminder, error)
	CreateReminder(userID uint, reminder *models.Reminder) error
	GetUserReminderByID(userID uint, reminderID uint) (*models.Reminder, error)
}

type reminderRepository struct {
	cache  map[uint][]models.Reminder
	dbLock sync.Mutex
}

func NewReminderRepository() ReminderRepository {
	return &reminderRepository{
		cache: make(map[uint][]models.Reminder),
	}
}

func (r *reminderRepository) GetUserReminders(userID uint) ([]models.Reminder, error) {
	r.dbLock.Lock()
	defer r.dbLock.Unlock()

	if reminders, ok := r.cache[userID]; ok {
		return reminders, nil
	}

	log.Printf("Reminders for user %d not found in cache, fetching from database...", userID)
	// В реальном приложении здесь должен быть вызов метода для получения напоминаний из базы данных

	return nil, nil // Placeholder
}

func (r *reminderRepository) GetUserReminderByID(userID uint, reminderID uint) (*models.Reminder, error) {
	r.dbLock.Lock()
	defer r.dbLock.Unlock()
	if reminders, ok := r.cache[userID]; ok {
		for _, reminder := range reminders {
			if reminder.ID == reminderID {
				return &reminder, nil
			}
		}
	}
	return nil, errors.New("reminder not found")
}

func (r *reminderRepository) CreateReminder(userID uint, reminder *models.Reminder) error {
	r.dbLock.Lock()
	defer r.dbLock.Unlock()

	log.Printf("Creating reminder for user %d: %v", userID, reminder)
	// В реальном приложении здесь должен быть вызов метода для создания напоминания в базе данных

	// Обновляем кэш
	if _, ok := r.cache[userID]; !ok {
		reminder.ID = 0
		r.cache[userID] = []models.Reminder{*reminder}
	} else {
		currID := len(r.cache[userID])
		reminder.ID = uint(currID)
		r.cache[userID] = append(r.cache[userID], *reminder)
	}

	return nil
}
