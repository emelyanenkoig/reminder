package bot

import (
	"bytes"
	"emelyanenkoig/reminder/pkg/models"
	"emelyanenkoig/reminder/pkg/repository"
	"encoding/json"
	"fmt"
	"gopkg.in/telebot.v3"
	"net/http"
	"strconv"
	"time"
)

type Bot struct {
	bot          *telebot.Bot
	userRepo     repository.UserRepository
	reminderRepo repository.ReminderRepository
}

func NewBot(token string, userRepo repository.UserRepository, reminderRepo repository.ReminderRepository) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	bot := &Bot{bot: b, userRepo: userRepo, reminderRepo: reminderRepo}

	//b.Handle("/start", bot.handleStart(userRepo, reminderRepo))
	//b.Handle("/add_reminder", bot.handleAddReminder)
	b.Handle("/add_user", bot.handleAddUser())
	b.Handle("/get_user", bot.handleGetUser())
	b.Handle("/add_reminder", bot.handleAddReminder())
	b.Handle("/get_reminders", bot.handleGetReminders())
	b.Handle("/get_reminder", bot.handleGetReminder())
	b.Handle("/update_reminder", bot.handleUpdateReminder())
	b.Handle("/delete_reminder", bot.handleDeleteReminder())

	return bot, nil
}

func (b *Bot) Start() {
	b.bot.Start()
}

func (b *Bot) handleStart(userRepo repository.UserRepository, reminderRepo repository.ReminderRepository) telebot.HandlerFunc {
	return func(c telebot.Context) error {
		//userId := c.Sender().ID
		//// getUserbyId(userID
		b.handleAddUser()
		return c.Send("Welcome! Your user profile has been created. You can now add reminders.")
	}
}

func (b *Bot) handleAddUser() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		// Отправка HTTP-запроса на получение пользователя
		response, err := http.Get(fmt.Sprintf("http://0.0.0.0:8080/%d", userID))
		if err != nil {
			return c.Send(err)
		}
		if response.StatusCode == http.StatusOK {
			return c.Send("User already exist" + response.Status)
		}

		newUser := &models.User{
			ID:        userID,
			Username:  c.Sender().Username,
			Reminders: make([]models.Reminder, 0),
		}

		jsonData, err := json.Marshal(newUser)
		if err != nil {
			return c.Send("Failed to marshal user data.")
		}

		response, err = http.Post("http://0.0.0.0:8080/", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return c.Send("Failed to create a new user.")
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusCreated {
			return c.Send("Failed to create a new user. Server responded with status: " + response.Status)
		}

		return c.Send(fmt.Sprintf("Your user profile has been created: %s.\n You can now add reminders.", newUser.Username))
	}
}

func (b *Bot) handleGetUser() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		// Отправка HTTP-запроса на получение пользователя
		response, err := http.Get(fmt.Sprintf("http://0.0.0.0:8080/%d", userID))
		if err != nil {
			return c.Send(err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return c.Send("User not found.")
		}

		var user models.User
		if err := json.NewDecoder(response.Body).Decode(&user); err != nil {
			return c.Send("Failed to decode user data.")
		}

		userInfo := fmt.Sprintf("User ID: %d\nUsername: %s\nReminders: %d", user.ID, user.Username, user.Reminders)
		return c.Send(userInfo)
	}
}

func (b *Bot) handleAddReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		args := c.Args()
		if len(args) < 2 {
			return c.Send("Please provide title.")
		}

		title := args[0]
		description := args[1]

		// Отправка HTTP-запроса на получение пользователя
		response, err := http.Get(fmt.Sprintf("http://0.0.0.0:8080/%d", userID))
		if err != nil {
			return c.Send(err)
		}
		if response.StatusCode != http.StatusOK {
			return c.Send("User is not exist" + response.Status)
		}

		newReminder := &models.Reminder{
			UserID:      userID,
			Title:       title,
			Description: description,
			DueDate:     time.Now().Add(time.Hour * 24),
		}

		jsonData, err := json.Marshal(newReminder)
		if err != nil {
			return c.Send("Failed to marshal reminder data.")
		}

		response, err = http.Post(fmt.Sprintf("http://0.0.0.0:8080/%d/reminders/", userID), "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return c.Send("Failed to create a new reminder.")
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusCreated {
			return c.Send("Failed to create a new reminder. Server responded with status: " + response.Status)
		}

		return c.Send(fmt.Sprintf("Your reminder has been created: %s.\n", newReminder.Title))
	}
}

func (b *Bot) handleGetReminders() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		// Отправка HTTP-запроса на получение Reminders
		response, err := http.Get(fmt.Sprintf("http://0.0.0.0:8080/%d/reminders", userID))
		if err != nil {
			return c.Send(err)
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return c.Send("User not found.")
		}

		var reminders []models.Reminder
		if err := json.NewDecoder(response.Body).Decode(&reminders); err != nil {
			return c.Send("Failed to decode user data.")
		}

		userInfo := fmt.Sprintf("Reminders %v", reminders)
		return c.Send(userInfo)
	}
}

func (b *Bot) handleGetReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		args := c.Args()
		fmt.Println(args)
		if len(args) == 0 {
			return c.Send("Please provide a reminder ID.")
		}

		userID := uint(c.Sender().ID)
		reminderID := args[0] // первый аргумент

		// Отправка HTTP-запроса на получение напоминания
		response, err := http.Get(fmt.Sprintf("http://0.0.0.0:8080/%d/reminders/%s", userID, reminderID))
		if err != nil {
			return c.Send("Failed to get reminder.")
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return c.Send("Reminder not found.")
		}

		var reminder models.Reminder
		if err := json.NewDecoder(response.Body).Decode(&reminder); err != nil {
			return c.Send("Failed to decode reminder data.")
		}

		// Формирование ответа для отправки пользователю
		reminderInfo := fmt.Sprintf("Reminder ID: %d\nTitle: %s\nDescription: %s", reminder.ID, reminder.Title, reminder.Description)
		return c.Send(reminderInfo)
	}
}

func (b *Bot) handleUpdateReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		args := c.Args()
		if len(args) < 3 {
			return c.Send("Please provide both reminder ID and new title.")
		}

		userID := uint(c.Sender().ID)
		strReminderId := args[0]
		reminderIDUint64, _ := strconv.Atoi(strReminderId)
		reminderID := uint(reminderIDUint64)

		newTitle := args[1]
		newDescription := args[2]

		updatedReminder := &models.Reminder{
			Title:       newTitle,
			Description: newDescription,
			DueDate:     time.Now().Add(time.Hour),
		}

		jsonData, err := json.Marshal(updatedReminder)
		if err != nil {
			return c.Send("Failed to marshal update data.")
		}
		client := &http.Client{}
		req, err := http.NewRequest("PUT", fmt.Sprintf("http://0.0.0.0:8080/%d/reminders/%d", userID, reminderID), bytes.NewBuffer(jsonData))
		if err != nil {
			return c.Send("Failed to create request.")
		}
		req.Header.Set("Content-Type", "application/json")

		response, err := client.Do(req)
		if err != nil {
			return c.Send("Failed to update reminder.")
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return c.Send("Failed to update reminder. Server responded with status: " + response.Status)
		}

		return c.Send("Reminder updated successfully.")
	}
}

func (b *Bot) handleDeleteReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		args := c.Args()
		if len(args) < 1 {
			return c.Send("Please provide a reminder ID.")
		}

		userID := uint(c.Sender().ID)
		reminderID := args[0]

		client := &http.Client{}
		req, err := http.NewRequest("DELETE", fmt.Sprintf("http://0.0.0.0:8080/%d/reminders/%s", userID, reminderID), nil)
		if err != nil {
			return c.Send("Failed to create request.")
		}

		response, err := client.Do(req)
		if err != nil {
			return c.Send("Failed to delete reminder.")
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return c.Send("Failed to delete reminder. Server responded with status: " + response.Status)
		}

		return c.Send("Reminder deleted successfully.")
	}
}
