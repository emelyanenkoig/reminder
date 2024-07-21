package models

import "time"

type User struct {
	ID        uint       `json:"id" binding:"required" gorm:"primaryKey"`
	Username  string     `json:"username" binding:"required"`
	Reminders []Reminder `json:"reminders" gorm:"foreignKey:UserID"`
}

type Reminder struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id" gorm:"index"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description" binding:"required"`
	DueDate     time.Time `json:"due_date" binding:"required"`
}
