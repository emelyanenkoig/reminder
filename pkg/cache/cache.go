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
	if exist {
		return user, true
	}
	return nil, false
}

func (c *Cache) GetUsers() []models.User {
	c.mu.RLock()
	defer c.mu.RUnlock()

	users := make([]models.User, 0, len(c.users))
	for _, user := range c.users {
		users = append(users, *user)
	}
	return users
}

func (c *Cache) DeleteUser(userId uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.users, userId)
}

func (c *Cache) AddReminder(userid uint, reminder *models.Reminder) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[userid].Reminders = append(c.users[userid].Reminders, *reminder)
}

func (c *Cache) GetReminderByUserId(userID uint, reminderID uint) (*models.Reminder, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if user, found := c.users[userID]; found {
		for _, reminder := range user.Reminders {
			if reminder.ID == reminderID {
				return &reminder, true
			}
		}
	}
	return nil, false
}

func (c *Cache) GetRemindersListByUser(userID uint) ([]models.Reminder, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if user, found := c.users[userID]; found {
		return user.Reminders, true
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
