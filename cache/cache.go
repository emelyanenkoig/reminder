package cache

import (
	"emelyanenkoig/reminder/models"
	"sync"
)

type Cache struct {
	mu    sync.Mutex
	cache map[uint]models.User
}

func NewCache() *Cache {
	return &Cache{
		cache: make(map[uint]models.User),
	}
}

func (c *Cache) AddUser(user models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[user.ID] = user
}

func (c *Cache) SetUser(userId uint, user models.User) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[userId] = user
}

func (c *Cache) GetUser(id uint) (models.User, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	user, exist := c.cache[id]
	return user, exist
}
