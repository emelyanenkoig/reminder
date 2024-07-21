package bot

import (
	"emelyanenkoig/reminder/pkg/models"
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"strconv"
	"time"
)

type UserState struct {
	State      string
	Date       string
	DateTime   string
	Title      string
	ReminderID int
}

const (
	StateChoosingDate     = "choosing_date"
	StateSettingDate      = "setting_date"
	StateSettingTime      = "setting_time"
	StateSettingTitle     = "setting_title"
	StateDeletingReminder = "deleting_reminder"
)

func (b *Bot) HandleStart() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)
		newUser := &models.User{
			ID:        userID,
			Username:  c.Sender().Username,
			Reminders: make([]models.Reminder, 0),
		}

		err := b.userService.CreateUser(newUser)
		if err != nil {
			log.Println("Error creating user:", err)
			return c.Send("Failed to create user: " + err.Error())
		}

		return c.Send(fmt.Sprintf("Your user profile has been created: %s.\nYou can now add reminders.", newUser.Username))
	}
}

func (b *Bot) HandleGetUser() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		user, err := b.userService.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			return c.Send("User not found.")
		}

		userInfo := fmt.Sprintf("User ID: %d\nUsername: %s\nReminders: %d", user.ID, user.Username, len(user.Reminders))
		return c.Send(userInfo)
	}
}

func (b *Bot) HandleAddReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		b.userStates[c.Chat().ID] = &UserState{State: StateSettingTitle}
		return c.Send("Please enter the title for the reminder.")
	}
}

func (b *Bot) HandleGetReminders() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		reminders, err := b.reminderService.GetRemindersByUserID(userID)
		if err != nil {
			log.Println("Error retrieving reminders:", err)
			return c.Send("Failed to retrieve reminders: " + err.Error())
		}

		if len(reminders) == 0 {
			return c.Send("You have no reminders.")
		}

		var remindersText string
		for _, reminder := range reminders {
			remindersText += fmt.Sprintf("ID: %d\nTitle: %s\nDescription: %s\nDue Date: %s\n\n", reminder.ID, reminder.Title, reminder.Description, reminder.DueDate.Format("2006-01-02 15:04"))
		}
		return c.Send(remindersText)
	}
}

func (b *Bot) HandleGetReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		args := c.Args()
		if len(args) < 1 {
			return c.Send("Please provide the reminder ID.")
		}

		reminderID, err := strconv.Atoi(args[0])
		if err != nil {
			return c.Send("Invalid reminder ID.")
		}

		reminder, err := b.reminderService.GetReminderByID(uint(reminderID), uint(c.Sender().ID))
		if err != nil {
			log.Println("Error retrieving reminder:", err)
			return c.Send("Reminder not found.")
		}

		reminderText := fmt.Sprintf("ID: %d\nTitle: %s\nDescription: %s\nDue Date: %s", reminder.ID, reminder.Title, reminder.Description, reminder.DueDate.Format("2006-01-02 15:04"))
		return c.Send(reminderText)
	}
}

func (b *Bot) HandleDeleteReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		args := c.Args()
		if len(args) < 1 {
			return c.Send("Please provide the reminder ID.")
		}

		reminderID, err := strconv.Atoi(args[0])
		if err != nil {
			return c.Send("Invalid reminder ID.")
		}

		userState := &UserState{
			State:      StateDeletingReminder,
			ReminderID: reminderID,
		}
		b.userStates[c.Chat().ID] = userState

		return c.Send(fmt.Sprintf("Are you sure you want to delete reminder with ID %d? (yes/no)", reminderID))
	}
}

func (b *Bot) HandleDeleteText() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		userState, ok := b.userStates[chatID]
		if !ok || userState.State != StateDeletingReminder {
			return c.Send("Please use /delete_reminder to start deleting a reminder.")
		}

		if c.Text() != "yes" {
			delete(b.userStates, chatID)
			return c.Send("Deletion cancelled.")
		}

		reminderID := userState.ReminderID
		err := b.reminderService.DeleteReminder(uint(c.Sender().ID), uint(reminderID))
		if err != nil {
			log.Println("Error deleting reminder:", err)
			return c.Send("Failed to delete reminder: " + err.Error())
		}

		delete(b.userStates, chatID)
		return c.Send(fmt.Sprintf("Reminder with ID %d has been deleted.", reminderID))
	}
}

func (b *Bot) HandleUpdateReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		args := c.Args()
		if len(args) < 1 {
			return c.Send("Please provide the reminder ID.")
		}

		reminderID, err := strconv.Atoi(args[0])
		if err != nil {
			return c.Send("Invalid reminder ID.")
		}

		_, err = b.reminderService.GetReminderByID(uint(reminderID), userID)
		if err != nil {
			log.Println("Error retrieving reminder:", err)
			return c.Send("Reminder not found.")
		}

		b.userStates[c.Chat().ID] = &UserState{
			State:      StateSettingTitle,
			ReminderID: reminderID,
		}
		return c.Send("Please enter the new title of the reminder.")
	}
}

func (b *Bot) HandleUpdateText() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		userID := uint(c.Sender().ID)
		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send("Unknown state. Please use /update_reminder to start updating a reminder.")
		}

		// Важно: Проверьте текущее состояние пользователя
		switch userState.State {
		case StateSettingTitle:
			newTitle := c.Text()
			reminder, err := b.reminderService.GetReminderByID(uint(userState.ReminderID), uint(c.Sender().ID))
			if err != nil {
				log.Println("Error retrieving reminder:", err)
				return c.Send("Reminder not found.")
			}

			reminder.Title = newTitle
			userState.Title = newTitle
			userState.State = StateChoosingDate

			err = b.reminderService.UpdateReminder(userID, uint(userState.ReminderID), reminder)
			if err != nil {
				log.Println("Error updating reminder:", err)
				return c.Send("Failed to update reminder: " + err.Error())
			}

			return c.Send("When would you like to set the reminder?", &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						{Text: "Today", Data: "today"},
						{Text: "Tomorrow", Data: "tomorrow"},
					},
					{
						{Text: "Set Date", Data: "set_date"},
					},
				},
			})

		case StateSettingDate:
			// Обработка установки даты и времени
			newDateStr := c.Text()
			newDate, err := time.Parse("2006-01-02", newDateStr)
			if err != nil {
				return c.Send("Invalid date format. Please use YYYY-MM-DD.")
			}
			userState.DateTime = newDate.Format("2006-01-02")
			userState.State = StateSettingTime
			return c.Send("Please enter the time (HH:MM).")

		case StateSettingTime:
			// Обработка установки времени
			newTimeStr := c.Text()
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Invalid time format. Please use HH:MM.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			reminder, err := b.reminderService.GetReminderByID(uint(userState.ReminderID), uint(c.Sender().ID))
			if err != nil {
				log.Println("Error retrieving reminder:", err)
				return c.Send("Reminder not found.")
			}

			newDueDateTime, err := time.Parse("2006-01-02 15:04", userState.DateTime)
			if err != nil {
				return c.Send("Failed to parse the due date and time.")
			}

			reminder.DueDate = newDueDateTime
			err = b.reminderService.UpdateReminder(userID, uint(userState.ReminderID), reminder)
			if err != nil {
				log.Println("Error updating reminder:", err)
				return c.Send("Failed to update reminder: " + err.Error())
			}

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Reminder updated with title: %s and due date: %s", reminder.Title, reminder.DueDate.Format("2006-01-02 15:04")))

		case StateDeletingReminder:
			chatID := c.Chat().ID
			userState, ok := b.userStates[chatID]
			if !ok || userState.State != StateDeletingReminder {
				return c.Send("Please use /delete_reminder to start deleting a reminder.")
			}

			if c.Text() != "yes" {
				delete(b.userStates, chatID)
				return c.Send("Deletion cancelled.")
			}

			reminderID := userState.ReminderID
			err := b.reminderService.DeleteReminder(uint(c.Sender().ID), uint(reminderID))
			if err != nil {
				log.Println("Error deleting reminder:", err)
				return c.Send("Failed to delete reminder: " + err.Error())
			}

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Reminder with ID %d has been deleted.", reminderID))

		default:
			return c.Send("Unknown state. Please use /update_reminder to start updating a reminder.")
		}
	}
}

// Обработка выбора даты и времени
func (b *Bot) HandleCallback() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		data := c.Callback().Data

		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send("Unknown callback data. Please use /update_reminder to start updating a reminder.")
		}

		switch data {
		case "today":
			today := time.Now().Format("2006-01-02")
			userState.DateTime = today
			userState.State = StateSettingTime
			return c.Send(fmt.Sprintf("Selected due date: %s. Please enter the time (HH:MM).", today))

		case "tomorrow":
			tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
			userState.DateTime = tomorrow
			userState.State = StateSettingTime
			return c.Send(fmt.Sprintf("Selected due date: %s. Please enter the time (HH:MM).", tomorrow))

		case "set_date":
			userState.State = StateSettingDate
			return c.Send("Please enter the due date (YYYY-MM-DD).")

		default:
			return c.Send("Unknown callback data. Please use /update_reminder to start updating a reminder.")
		}
	}
}

// Установка новой даты и времени
func (b *Bot) HandleText() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send("Unknown state. Please use /update_reminder to start updating a reminder.")
		}

		switch userState.State {
		case StateSettingDate:
			newDateStr := c.Text()
			newDate, err := time.Parse("2006-01-02", newDateStr)
			if err != nil {
				return c.Send("Invalid date format. Please use YYYY-MM-DD.")
			}
			userState.DateTime = newDate.Format("2006-01-02")
			userState.State = StateSettingTime
			return c.Send("Please enter the time (HH:MM).")

		case StateSettingTime:
			newTimeStr := c.Text()
			userID := uint(c.Sender().ID)
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Invalid time format. Please use HH:MM.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			reminder, err := b.reminderService.GetReminderByID(uint(userState.ReminderID), uint(c.Sender().ID))
			if err != nil {
				log.Println("Error retrieving reminder:", err)
				return c.Send("Reminder not found.")
			}

			newDueDateTime, err := time.Parse("2006-01-02 15:04", userState.DateTime)
			if err != nil {
				return c.Send("Failed to parse the due date and time.")
			}

			reminder.DueDate = newDueDateTime
			err = b.reminderService.UpdateReminder(userID, uint(userState.ReminderID), reminder)
			if err != nil {
				log.Println("Error updating reminder:", err)
				return c.Send("Failed to update reminder: " + err.Error())
			}

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Reminder updated with title: %s and due date: %s", reminder.Title, reminder.DueDate.Format("2006-01-02 15:04")))

		default:
			return c.Send("Unknown state. Please use /update_reminder to start updating a reminder.")
		}
	}
}
