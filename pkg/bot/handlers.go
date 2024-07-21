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
	StateChoosingDate            = "choosing_date"
	StateCreatingTitle           = "creating_title"
	StateCreatingDate            = "creating_date"
	StateCreatingTime            = "creating_time"
	StateCreatingRepositoryModel = "creating_repository_model"

	StateSettingDate  = "setting_date"
	StateSettingTime  = "setting_time"
	StateSettingTitle = "setting_title"

	StateDeletingReminder = "deleting_reminder"
)

// Обработчик команды /start
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

// Обработчик команды /get_user
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

// Обработчик команды /add_reminder
func (b *Bot) HandleAddReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		b.userStates[c.Chat().ID] = &UserState{State: StateCreatingTitle}
		return c.Send("Please enter the title for the reminder.")
	}
}

// Обработчик команды /get_reminders
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

// Обработчик команды /get_reminder
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

// Обработчик команды /delete_reminder
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

// Обработчик подтверждения удаления напоминания
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

// Обработчик команды /update_reminder
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

		switch userState.State {
		case StateCreatingTitle:
			userState.Title = c.Text()
			userState.State = StateCreatingDate
			return c.Send("When would you like to set the reminder?", &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						{Text: "Today", Data: "today"},
						{Text: "Tomorrow", Data: "tomorrow"},
						{Text: "Set Date", Data: "set_date"},
					},
				},
			})
		case StateCreatingDate:
			newDateStr := c.Text()
			newDate, err := time.Parse("2006-01-02", newDateStr)
			if err != nil {
				return c.Send("Invalid date format. Please use YYYY-MM-DD.")
			}
			userState.DateTime = newDate.Format("2006-01-02")
			return c.Send("Please enter the time (HH:MM).", &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						{Text: "09:00", Data: "09:00"},
						{Text: "12:00", Data: "12:00"},
						{Text: "15:00", Data: "15:00"},
						{Text: "18:00", Data: "18:00"},
						{Text: "21:00", Data: "21:00"},
						{Text: "set time", Data: "set_time"},
					},
				},
				ResizeKeyboard: true,
			})
		case StateCreatingTime:
			newTimeStr := c.Text()
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Invalid time format. Please use HH:MM.")
			}
			userState.State = StateCreatingRepositoryModel
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)
			return c.Send(fmt.Sprintf("Title: %s \nDueDate: %v\n", userState.Title, userState.DateTime), &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						{Text: "YES", Data: "create_YES_reminder"},
						{Text: "NO", Data: "create_NO_reminder"},
					},
				},
			})
		case StateCreatingRepositoryModel:
			dueDateTime, err := time.Parse("2006-01-02 15:04", userState.DateTime)
			reminder := &models.Reminder{
				UserID:  userID,
				Title:   userState.Title,
				DueDate: dueDateTime,
			}

			err = b.reminderService.CreateReminder(reminder)
			if err != nil {
				log.Println("Error creating reminder:", err)
				return c.Send("Failed to create reminder: " + err.Error())
			}
			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Reminder created with title: %s and due date: %s", reminder.Title, reminder.DueDate.Format("2006-01-02 15:04")))

		//case StateSettingTitle:
		//	newTitle := c.Text()
		//	reminder, err := b.reminderService.GetReminderByID(uint(userState.ReminderID), userID)
		//	if err != nil {
		//		log.Println("Error retrieving reminder:", err)
		//		return c.Send("Reminder not found.")
		//	}
		//
		//	reminder.Title = newTitle
		//	userState.Title = newTitle
		//	userState.State = StateChoosingDate
		//
		//	err = b.reminderService.UpdateReminder(userID, uint(userState.ReminderID), reminder)
		//	if err != nil {
		//		log.Println("Error updating reminder:", err)
		//		return c.Send("Failed to update reminder: " + err.Error())
		//	}
		//
		//	return c.Send("When would you like to set the reminder?", &telebot.ReplyMarkup{
		//		InlineKeyboard: [][]telebot.InlineButton{
		//			{
		//				{Text: "Today", Data: "today"},
		//				{Text: "Tomorrow", Data: "tomorrow"},
		//				{Text: "Set Date", Data: "set_date"},
		//			},
		//		},
		//	})
		//case StateSettingDate:
		//	newDateStr := c.Text()
		//	newDate, err := time.Parse("2006-01-02", newDateStr)
		//	if err != nil {
		//		return c.Send("Invalid date format. Please use YYYY-MM-DD.")
		//	}
		//	userState.DateTime = newDate.Format("2006-01-02")
		//	userState.State = StateSettingTime
		//	return c.Send("Please enter the time (HH:MM).")
		//case StateSettingTime:
		//	newTimeStr := c.Text()
		//	_, err := time.Parse("15:04", newTimeStr)
		//	if err != nil {
		//		return c.Send("Invalid time format. Please use HH:MM.")
		//	}
		//	userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)
		//	dueDateTime, err := time.Parse("2006-01-02 15:04", userState.DateTime)
		//	if err != nil {
		//		return c.Send("Failed to parse the due date and time.")
		//	}
		//
		//	reminder, err := b.reminderService.GetReminderByID(uint(userState.ReminderID), userID)
		//	if err != nil {
		//		log.Println("Error retrieving reminder:", err)
		//		return c.Send("Reminder not found.")
		//	}
		//
		//	reminder.DueDate = dueDateTime
		//	err = b.reminderService.UpdateReminder(userID, uint(userState.ReminderID), reminder)
		//	if err != nil {
		//		log.Println("Error updating reminder:", err)
		//		return c.Send("Failed to update reminder: " + err.Error())
		//	}
		//
		//	delete(b.userStates, chatID)
		//	return c.Send(fmt.Sprintf("Reminder updated with title: %s and due date: %s", reminder.Title, reminder.DueDate.Format("2006-01-02 15:04")))

		default:
			return c.Send("Unknown state. Please use /update_reminder to start updating a reminder.")
		}
	}
}

func (b *Bot) HandleText() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		userID := uint(c.Sender().ID)
		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send("Unknown state. Please use /add_reminder to start creating a reminder.")
		}

		switch userState.State {
		case StateCreatingTitle:
			title := c.Text()
			userState.Title = title
			userState.State = StateSettingDate
			return c.Send("Please enter the date for the reminder (YYYY-MM-DD).")

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
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Invalid time format. Please use HH:MM.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)
			dueDateTime, err := time.Parse("2006-01-02 15:04", userState.DateTime)
			if err != nil {
				return c.Send("Failed to parse the due date and time.")
			}

			reminder := models.Reminder{
				UserID:      userID,
				Title:       userState.Title,
				Description: "",
				DueDate:     dueDateTime,
			}

			err = b.reminderService.CreateReminder(&reminder)
			if err != nil {
				log.Println("Error creating reminder:", err)
				return c.Send("Failed to create reminder: " + err.Error())
			}

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Reminder created with title: %s and due date: %s", reminder.Title, reminder.DueDate.Format("2006-01-02 15:04")))

		default:
			return c.Send("Unknown state. Please use /add_reminder to start creating a reminder.")
		}
	}
}
