package models

import "time"

type User struct {
	ID        uint       `json:"id" binding:"required"`
	Username  string     `json:"username" binding:"required"`
	Reminders []Reminder `json:"reminders"`
}

func NewUser(id uint, username string, reminders []Reminder) *User {
	return &User{
		ID:        id,
		Username:  username,
		Reminders: reminders,
	}
}

type Reminder struct {
	ID          uint      `json:"-"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
	UpdatedAt   time.Time `json:"-"`
	CreatedAt   time.Time `json:"-"`
}

func NewReminder(id uint, title string, dueDate, createdAt, updatedAt time.Time) *Reminder {
	return &Reminder{
		ID:        id,
		Title:     title,
		DueDate:   dueDate,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
