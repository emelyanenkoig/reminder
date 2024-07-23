package bot

import (
	"emelyanenkoig/reminder/pkg/models"
	"fmt"
	"gopkg.in/telebot.v3"
	"log"
	"strconv"
	"strings"
	"time"
)

type UserState struct {
	State      string
	DateTime   string
	Title      string
	ReminderID int
}

const (
	StateCreatingTitle    = "creating_title"
	StateSettingDate      = "setting_date"
	StateSettingTime      = "setting_time"
	StateDeletingReminder = "deleting_reminder"

	StateUpdatingReminderTitle = "updating_reminder_title"
	StateUpdatingReminderDate  = "updating_reminder_date"
	StateUpdatingReminderTime  = "updating_reminder_time"
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
			return c.Send("Не удалось создать пользователя: " + err.Error())
		}

		return c.Send(fmt.Sprintf("Ваш профиль создан: %s.\nТеперь вы можете добавлять напоминания.", newUser.Username))
	}
}

func (b *Bot) HandleGetUser() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		user, err := b.userService.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			return c.Send("Пользователь не найден.")
		}

		userInfo := fmt.Sprintf("ID пользователя: %d\nИмя пользователя: %s\nНапоминаний: %d", user.ID, user.Username, len(user.Reminders))
		return c.Send(userInfo)
	}
}

func (b *Bot) HandleAddReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		b.userStates[c.Chat().ID] = &UserState{State: StateCreatingTitle}
		return c.Send("Введите название напоминания:")
	}
}

func (b *Bot) HandleUpdateReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		user, err := b.userService.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			return c.Send("Пользователь не найден.")
		}

		if len(user.Reminders) == 0 {
			return c.Send("У вас нет напоминаний для обновления.")
		}

		b.userStates[c.Chat().ID] = &UserState{State: StateUpdatingReminderTitle}
		var buttons [][]telebot.InlineButton

		for _, reminder := range user.Reminders {
			buttons = append(buttons, []telebot.InlineButton{
				{Text: reminder.Title, Data: fmt.Sprintf("update_%d", reminder.ID)},
			})
		}

		return c.Send("Выберите напоминание для обновления:", &telebot.ReplyMarkup{
			InlineKeyboard: buttons,
		})
	}
}

func (b *Bot) HandleText() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		userID := uint(c.Sender().ID)
		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send("Неизвестное состояние. Используйте /add для создания напоминания.")
		}

		switch userState.State {
		case StateCreatingTitle:
			userState.Title = c.Text()
			userState.State = StateSettingDate
			return c.EditOrSend("Когда вы хотите установить напоминание?", &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{{Text: "Сегодня", Data: "today"}, {Text: "Завтра", Data: "tomorrow"}, {Text: "Установить дату", Data: "set_date"}},
				},
			})
		case StateSettingDate:
			dateStr := c.Text()
			_, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return c.Send("Неверный формат даты. Пожалуйста, используйте ГГГГ-ММ-ДД.")
			}
			userState.DateTime = dateStr
			userState.State = StateSettingTime
			return c.Send("Выберите время из предложенных вариантов или введите свое.", &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						{Text: "🌅 09:00", Data: "09:00"},
						{Text: "☀️ 12:00", Data: "12:00"},
						{Text: "☀️ 15:00", Data: "15:00"},
						{Text: "🌆 18:00", Data: "18:00"},
						{Text: "🌃 21:00", Data: "18:00"}},
				},
			})
		case StateSettingTime:
			newTimeStr := c.Text()
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Неверный формат времени. Пожалуйста, используйте ЧЧ:ММ.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			dueDateTime, err := time.Parse("2006-01-02 15:04", userState.DateTime)
			if err != nil {
				return c.Send("Не удалось разобрать дату и время.")
			}

			dueDateTime = dueDateTime.UTC()

			reminder := &models.Reminder{
				UserID:  userID,
				Title:   userState.Title,
				DueDate: dueDateTime,
			}

			err = b.reminderService.CreateReminder(reminder)
			if err != nil {
				log.Println("Error creating reminder:", err)
				return c.Send("Не удалось создать напоминание: " + err.Error())
			}

			// Запланируйте напоминание
			b.scheduleReminder(reminder)

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Напоминание создано:\n%s", reminder.Title))
		case StateUpdatingReminderTitle:
			newTitle := c.Text()
			userState.Title = newTitle
			userState.State = StateUpdatingReminderDate
			return c.Send("Введите новую дату для напоминания (ГГГГ-ММ-ДД):")
		case StateUpdatingReminderDate:
			newDate := c.Text()
			_, err := time.Parse("2006-01-02", newDate)
			if err != nil {
				return c.Send("Неверный формат даты. Пожалуйста, используйте ГГГГ-ММ-ДД.")
			}
			userState.DateTime = newDate
			userState.State = StateUpdatingReminderTime
			return c.Send("Введите новое время для напоминания (ЧЧ:ММ):")
		case StateUpdatingReminderTime:
			newTime := c.Text()
			_, err := time.Parse("15:04", newTime)
			if err != nil {
				return c.Send("Неверный формат времени. Пожалуйста, используйте ЧЧ:ММ.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTime)

			// Преобразование строки в время в UTC
			dueDateTime, err := time.Parse("2006-01-02 15:04", userState.DateTime)
			if err != nil {
				return c.Send("Не удалось разобрать дату и время.")
			}

			// Преобразуем в UTC
			dueDateTime = dueDateTime.UTC()

			reminder := &models.Reminder{
				ID:      uint(userState.ReminderID),
				UserID:  userID,
				Title:   userState.Title,
				DueDate: dueDateTime,
			}

			err = b.reminderService.UpdateReminder(userID, reminder.ID, reminder)
			if err != nil {
				log.Println("Error updating reminder:", err)
				return c.Send("Не удалось обновить напоминание: " + err.Error())
			}

			// Запланируйте обновленное напоминание
			b.scheduleReminder(reminder)

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Напоминание обновлено 🔄:\n%s", reminder.Title))

		default:
			return c.Send("Неизвестное состояние. Используйте /add для создания напоминания.")
		}
	}
}

func (b *Bot) HandleCallback() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		data := c.Callback().Data

		if data == "" {
			return c.Send("Неизвестная ошибка. Попробуйте снова.")
		}

		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send("Неизвестное состояние. Используйте /add для создания напоминания.")
		}

		switch userState.State {
		case StateSettingDate:
			var dateStr string
			switch data {
			case "today":
				dateStr = time.Now().Format("2006-01-02")
			case "tomorrow":
				dateStr = time.Now().AddDate(0, 0, 1).Format("2006-01-02")
			case "set_date":
				userState.State = StateSettingDate
				return c.Send("Введите дату (ГГГГ-ММ-ДД).")
			default:
				return c.Send("Неизвестный выбор. Пожалуйста, попробуйте снова.")
			}
			userState.DateTime = dateStr
			userState.State = StateSettingTime
			return c.EditOrSend("Выберите время из предложенных вариантов или введите свое.", &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						{Text: "🌅 09:00", Data: "09:00"},
						{Text: "☀️ 12:00", Data: "12:00"},
						{Text: "☀️ 15:00", Data: "15:00"},
						{Text: "🌆 18:00", Data: "18:00"},
						{Text: "🌃 21:00", Data: "18:00"},
					},
				}, OneTimeKeyboard: true,
			})
		case StateSettingTime:
			newTimeStr := data
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Неверный формат времени. Пожалуйста, используйте ЧЧ:ММ.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			dueDateTime, err := time.ParseInLocation("2006-01-02 15:04", userState.DateTime, time.UTC)
			if err != nil {
				return c.Send("Не удалось разобрать дату и время.")
			}

			reminder := &models.Reminder{
				UserID:  uint(c.Sender().ID),
				Title:   userState.Title,
				DueDate: dueDateTime,
			}

			err = b.reminderService.CreateReminder(reminder)
			if err != nil {
				log.Println("Error creating reminder:", err)
				return c.Send("Не удалось создать напоминание: " + err.Error())
			}

			// Запланируйте напоминание
			b.scheduleReminder(reminder)

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Напоминание создано 🤝:\n%s", reminder.Title))
		case StateDeletingReminder:
			if strings.HasPrefix(data, "delete_") {
				userID := uint(c.Sender().ID)
				reminderIDStr := strings.TrimPrefix(data, "delete_")
				reminderID, err := strconv.Atoi(reminderIDStr)
				if err != nil {
					return c.Send("Ошибка при разборе ID напоминания.")
				}

				err = b.reminderService.DeleteReminder(userID, uint(reminderID))
				if err != nil {
					return c.Send("Не удалось удалить напоминание: " + err.Error())
				}

				delete(b.userStates, chatID)
				return c.Send("♻♻♻")
			}
		case StateUpdatingReminderTitle:
			reminderIDStr := strings.TrimPrefix(data, "update_")
			reminderID, err := strconv.Atoi(reminderIDStr)
			if err != nil {
				return c.Send("Ошибка при разборе ID напоминания.")
			}

			userState.State = StateUpdatingReminderTitle
			userState.ReminderID = reminderID

			return c.Send("🔄 Введите новое название напоминания:")
		default:
			return c.Send("Неизвестное состояние. Используйте /add для создания напоминания.")
		}
		return nil
	}
}

// Обработчик команды /list
func (b *Bot) HandleListReminders() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		user, err := b.userService.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			return c.Send("Пользователь не найден.")
		}

		if len(user.Reminders) == 0 {
			return c.Send("У вас нет напоминаний.")
		}

		var buttons [][]telebot.InlineButton
		for _, reminder := range user.Reminders {
			buttons = append(buttons, []telebot.InlineButton{
				{Text: reminder.Title},
			})
		}

		return c.Send("📚 Список ваших напоминаний:", &telebot.ReplyMarkup{
			InlineKeyboard: buttons,
			RemoveKeyboard: true,
			ResizeKeyboard: true,
		})
	}
}

func (b *Bot) HandleDeleteReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		user, err := b.userService.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			return c.Send("Пользователь не найден.")
		}

		if len(user.Reminders) == 0 {
			return c.Send("У вас нет напоминаний для удаления.")
		}

		var buttons [][]telebot.InlineButton
		for _, reminder := range user.Reminders {
			buttons = append(buttons, []telebot.InlineButton{
				{Text: reminder.Title, Data: fmt.Sprintf("delete_%d", reminder.ID)},
			})
		}

		b.userStates[c.Chat().ID] = &UserState{State: StateDeletingReminder}

		return c.Send(" ♻ Выберите напоминание для удаления:", &telebot.ReplyMarkup{
			InlineKeyboard: buttons,
		})
	}
}

func (b *Bot) sendReminder(reminder *models.Reminder) {
	message := fmt.Sprintf("Напоминание: %s\nДата и время: %s", reminder.Title, reminder.DueDate.In(time.Local).Format("2006-01-02 15:04"))
	chatID := int64(reminder.UserID)

	_, err := b.Bot.Send(&telebot.Chat{ID: chatID}, message)
	if err != nil {
		log.Println("Error sending reminder:", err)
	}
}

func (b *Bot) scheduleReminder(reminder *models.Reminder) {
	duration := time.Until(reminder.DueDate.UTC())
	if duration <= 0 {
		log.Println("Reminder time is in the past")
		return
	}

	time.AfterFunc(duration, func() {
		b.sendReminder(reminder)
	})
}
