package models

import "time"

type User struct {
	ID        uint       `json:"id" binding:"required"`
	Username  string     `json:"username" binding:"required"`
	Reminders []Reminder `json:"reminders"`
}

type Reminder struct {
	ID          uint      `json:"-"`
	UserID      uint      `json:"-"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
	UpdatedAt   time.Time `json:"-"`
	CreatedAt   time.Time `json:"-"`
}
