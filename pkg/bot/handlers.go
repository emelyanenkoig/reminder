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
	StateViewingReminder  = "viewing_reminder"

	StateUpdatingReminderTitle = "updating_reminder_title"
	StateUpdatingReminderDate  = "updating_reminder_date"
	StateUpdatingReminderTime  = "updating_reminder_time"

	StandardError = "❗️*Неизвестное состояние* ❗\nИспользуйте /add для создания напоминания"
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
		return c.Send("⏰ Введите название напоминания:")
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
			return c.Send(StandardError, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}

		switch userState.State {
		case StateCreatingTitle:
			userState.Title = c.Text()
			userState.State = StateSettingDate
			return c.EditOrSend("Когда вы хотите установить напоминание?", &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{{Text: "Сегодня", Data: "today"}, {Text: "Завтра", Data: "tomorrow"}},
					{{Text: "Установить дату", Data: "set_date"}},
				},
				RemoveKeyboard: true,
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
					{{Text: "🌅 09:00", Data: "09:00"}, {Text: "☀️ 12:00", Data: "12:00"}},
					{{Text: "☀️ 15:00", Data: "15:00"}, {Text: "🌆 18:00", Data: "18:00"}},
					{{Text: "🌃 21:00", Data: "21:00"}},
				},
				RemoveKeyboard: true,
				ResizeKeyboard: true,
			})
		case StateSettingTime:
			newTimeStr := c.Text()
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Неверный формат времени. Пожалуйста, используйте ЧЧ:ММ.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			// Преобразуйте строку в время в местном времени
			dueDateTime, err := time.ParseInLocation("2006-01-02 15:04", userState.DateTime, time.Local)
			if err != nil {
				return c.Send("Не удалось разобрать дату и время.")
			}

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

			// Преобразование строки в время в местном времени
			dueDateTime, err := time.ParseInLocation("2006-01-02 15:04", userState.DateTime, time.Local)
			if err != nil {
				return c.Send("Не удалось разобрать дату и время.")
			}

			reminder := &models.Reminder{
				ID:      uint(userState.ReminderID),
				UserID:  userID,
				Title:   userState.Title,
				DueDate: dueDateTime,
			}

			err = b.reminderService.UpdateReminder(userID, reminder.ID, reminder) //
			if err != nil {
				log.Println("Error updating reminder:", err)
				return c.Send("Не удалось обновить напоминание: " + err.Error())
			}

			// Запланируйте обновленное напоминание
			b.scheduleReminder(reminder)

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf("Напоминание обновлено 🔄:\n%s", reminder.Title))

		default:
			return c.Send(StandardError, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}
	}
}

func (b *Bot) HandleCallback() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		chatID := c.Chat().ID
		data := c.Callback().Data
		userID := uint(c.Sender().ID)

		if data == "" {
			return c.Send("Неизвестная ошибка. Попробуйте снова.")
		}

		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send(StandardError, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}

		loc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			log.Println("Error loading location:", err)
			return c.Send("Не удалось загрузить временную зону.")
		}

		switch userState.State {
		case StateSettingDate:
			var dateStr string
			switch data {
			case "today":
				dateStr = time.Now().In(loc).Format("2006-01-02")
			case "tomorrow":
				dateStr = time.Now().In(loc).AddDate(0, 0, 1).Format("2006-01-02")
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
					{{Text: "🌅 09:00", Data: "09:00"}, {Text: "☀️ 12:00", Data: "12:00"}},
					{{Text: "☀️ 15:00", Data: "15:00"}, {Text: "🌆 18:00", Data: "18:00"}},
					{{Text: "🌃 21:00", Data: "21:00"}},
				},
				RemoveKeyboard: true,
				ResizeKeyboard: true,
			})
		case StateSettingTime:
			newTimeStr := data
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send("Неверный формат времени. Пожалуйста, используйте ЧЧ:ММ.")
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			dueDateTime, err := time.ParseInLocation("2006-01-02 15:04", userState.DateTime, loc)
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
		case StateViewingReminder:
			if strings.HasPrefix(data, "view_") {
				reminderIDStr := strings.TrimPrefix(data, "view_")
				reminderID, err := strconv.Atoi(reminderIDStr)
				if err != nil {
					return c.Send("Ошибка при разборе ID напоминания.")
				}

				reminder, err := b.reminderService.GetReminderByID(uint(reminderID), userID)
				if err != nil {
					return c.Send("Не удалось найти напоминание.")
				}

				message := fmt.Sprintf("📚 <b>Напоминание:</b> %s\n\n⏰ <b>Напомню Вам:</b> %s", reminder.Title, reminder.DueDate.In(time.Local).Format("2006-01-02 15:04"))
				return c.Send(message, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
			}
		default:
			return c.Send(StandardError, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}
		return nil
	}
}

// Обработчик команды /list
func (b *Bot) HandleListReminders() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)
		chatID := c.Chat().ID

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
				{Text: reminder.Title, Data: fmt.Sprintf("view_%d", reminder.ID)},
			})
		}

		// Устанавливаем состояние пользователя для просмотра напоминаний
		b.userStates[chatID] = &UserState{State: StateViewingReminder}

		return c.Send("📚 Список ваших напоминаний:", &telebot.ReplyMarkup{
			InlineKeyboard: buttons,
			RemoveKeyboard: true,
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
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Println("Error loading location:", err)
		return
	}

	reminder.DueDate = reminder.DueDate.In(loc)

	duration := time.Until(reminder.DueDate)
	if duration <= 0 {
		log.Println("Reminder time is in the past")
		return
	}

	time.AfterFunc(duration, func() {
		b.sendReminder(reminder)
	})
}
