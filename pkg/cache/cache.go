package cache

import (
	"emelyanenkoig/reminder/pkg/models"
	"sync"
)

type Cache struct {
	users map[uint]*models.User
	mu    sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		users: make(map[uint]*models.User),
	}
}

func (c *Cache) AddUser(user *models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[user.ID] = user
}

func (c *Cache) GetUser(id uint) (*models.User, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	user, exist := c.users[id]
	return user, exist
}

func (c *Cache) AddReminder(userid uint, reminder *models.Reminder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[userid].Reminders = append(c.users[userid].Reminders, *reminder)
}

func (c *Cache) GetReminderByUserId(reminderId uint, userId uint) (*models.Reminder, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	user, found := c.GetUser(userId)
	if found {
		for _, reminder := range user.Reminders {
			if reminder.ID == reminderId {
				return &reminder, true
			}
		}
	}
	return nil, false

}

func (c *Cache) GetRemindersListByUser(userId uint) ([]models.Reminder, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	user, found := c.GetUser(userId)
	if found {
		return user.Reminders, false
	}
	return nil, false

}

func (c *Cache) DeleteReminder(userId uint, reminderId uint) {
	c.mu.Lock()
	defer c.mu.Unlock()

	user, found := c.users[userId]
	if found {
		for i, reminder := range user.Reminders {
			if reminder.ID == reminderId {
				user.Reminders = append(user.Reminders[:i], user.Reminders[i+1:]...)
				return
			}
		}
	}
}
