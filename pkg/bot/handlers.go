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
)

const (
	StandardError                   = "❗️*Неизвестное состояние* ❗\n\n👁👄👁\nИспользуйте /add для создания напоминания"
	ErrorCallbackDataNull           = "Неизвестная ошибка. Попробуйте снова."
	ErrorCreateUser                 = "❗ *Не удалось создать пользователя:* ❗\n"
	ErrorGetUser                    = "❗ *Пользователь не найден* ❗"
	ErrorInvalidData                = "❗ *Неверный формат даты* ❗\nПожалуйста, используйте _ГГГГ\\-ММ\\-ДД_"
	ErrorInvalidTime                = "❗ *Неверный формат времени* ❗\n Пожалуйста, используйте _ЧЧ:ММ_"
	ErrorInvalidDateTimeCompilation = "❗ *Неверный формат даты и времени* ❗"
	ErrorCreateReminder             = "❗ *Не удалось создать напоминание* ❗"
	ErrorUpdateReminder             = "❗ *Не удалось обновить напоминание* ❗"
	ErrorLoadTimezone               = "❗ *Не удалось загрузить временную зону* ❗"
	ErrorParseIdOfReminder          = "❗ *Ошибка при разборе ID напоминания* ❗"
	ErrorFindReminder               = "❗ *Не удалось удалить напоминание* ❗"
)

const (
	MessageCreateUserSuccess       = "🗄 *Ваш профиль создан:* %s \n\n_Теперь вы можете добавлять напоминания_"
	MessageCreateReminderSuccess   = "✅ *Напоминание создано:*\n%s"
	MessageUpdateReminderSuccess   = "✅ *Напоминание обновлено:*\n%s"
	MessageInfoUser                = "ID пользователя: %d\nИмя пользователя: %s\nНапоминаний: %d"
	MessageEnterNameOfReminder     = "📌 *Введите название напоминания:*"
	MessageEnterNewNameOfReminder  = "📌 *Введите новое название напоминания:*"
	MessageNoSuchReminders         = "У вас нет напоминаний"
	MessageChooseReminderForUpdate = "🔎 *Выберите напоминание для обновления:*"
	MessageChooseReminderForDelete = " ♻ *Выберите напоминание для удаления:*"
	MessageChooseDateOfReminder    = "📅 *Когда вы хотите установить напоминание?*"
	MessageSetDateOfReminder       = "📅 *Введите дату* _ГГГГ\\-ММ\\-ДД_:"
	MessageSetNewDateOfReminder    = "📅 *Введите новую дату для напоминания* _ГГГГ\\-ММ\\-ДД_:"
	MessageChooseTimeOfReminder    = "🕰 *Выберите время из предложенных вариантов или введите свое*"
	MessageChooseNewTimeOfReminder = "🕰 *Введите новое время для напоминания* _ЧЧ:ММ_:"
	MessagePrintReminderData       = "📚 Напоминание: %s\n\n⏰ Напомню Вам: %s"
	MessageListReminders           = "📚 *Список ваших напоминаний*"
	MessageReminderUser            = "⏰ Напоминание: %s\nДата и время: %s"
)

const (
	KeyboardToday    = "Сегодня"
	KeyboardTomorrow = "Завтра"
	KeyboardSetDate  = "Установить дату"
	KeyboardTime9    = "🌅 09:00"
	KeyboardTime12   = "☀️ 12:00"
	KeyboardTime15   = "☀️ 15:00"
	KeyboardTime18   = "🌆 18:00"
	KeyboardTime21   = "🌃 21:00"
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
			return c.Send(ErrorCreateUser + err.Error())
		}

		return c.Send(fmt.Sprintf(MessageCreateUserSuccess, newUser.Username), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
	}
}

func (b *Bot) HandleGetUser() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		user, err := b.userService.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			return c.Send(ErrorGetUser, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}

		userInfo := fmt.Sprintf(MessageInfoUser, user.ID, user.Username, len(user.Reminders))
		return c.Send(userInfo, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
	}
}

func (b *Bot) HandleAddReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		b.userStates[c.Chat().ID] = &UserState{State: StateCreatingTitle}
		return c.Send(MessageEnterNameOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
	}
}

func (b *Bot) HandleUpdateReminder() telebot.HandlerFunc {
	return func(c telebot.Context) error {
		userID := uint(c.Sender().ID)

		user, err := b.userService.GetUserByID(userID)
		if err != nil {
			log.Println("Error getting user:", err)
			return c.Send(ErrorGetUser, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}

		if len(user.Reminders) == 0 {
			return c.Send(MessageNoSuchReminders)
		}

		b.userStates[c.Chat().ID] = &UserState{State: StateUpdatingReminderTitle}
		var buttons [][]telebot.InlineButton

		for _, reminder := range user.Reminders {
			buttons = append(buttons, []telebot.InlineButton{
				{Text: reminder.Title, Data: fmt.Sprintf("update_%d", reminder.ID)},
			})
		}

		return c.Send(MessageChooseReminderForUpdate, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2}, &telebot.ReplyMarkup{
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
			return c.EditOrSend(MessageChooseDateOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2}, &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{{Text: KeyboardToday, Data: "today"}, {Text: KeyboardTomorrow, Data: "tomorrow"}},
					{{Text: KeyboardSetDate, Data: "set_date"}},
				},
				RemoveKeyboard: true,
			})
		case StateSettingDate:
			dateStr := c.Text()
			_, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				return c.Send(ErrorInvalidData, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}
			userState.DateTime = dateStr
			userState.State = StateSettingTime
			return c.Send(MessageChooseTimeOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2}, &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{{Text: KeyboardTime9, Data: "09:00"}, {Text: KeyboardTime12, Data: "12:00"}},
					{{Text: KeyboardTime15, Data: "15:00"}, {Text: KeyboardTime18, Data: "18:00"}},
					{{Text: KeyboardTime21, Data: "21:00"}},
				},
				RemoveKeyboard: true,
				ResizeKeyboard: true,
			})
		case StateSettingTime:
			newTimeStr := c.Text()
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send(ErrorInvalidTime, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			dueDateTime, err := time.ParseInLocation("2006-01-02 15:04", userState.DateTime, time.Local)
			if err != nil {
				return c.Send(ErrorInvalidDateTimeCompilation, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}

			reminder := &models.Reminder{
				UserID:  userID,
				Title:   userState.Title,
				DueDate: dueDateTime,
			}

			err = b.reminderService.CreateReminder(reminder)
			if err != nil {
				log.Println("Error creating reminder:", err)
				return c.Send(ErrorCreateReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}

			b.scheduleReminder(reminder)

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf(MessageCreateReminderSuccess, reminder.Title), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		case StateUpdatingReminderTitle:
			newTitle := c.Text()
			userState.Title = newTitle
			userState.State = StateUpdatingReminderDate
			return c.Send(MessageSetNewDateOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		case StateUpdatingReminderDate:
			newDate := c.Text()
			_, err := time.Parse("2006-01-02", newDate)
			if err != nil {
				return c.Send(ErrorInvalidData, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}
			userState.DateTime = newDate
			userState.State = StateUpdatingReminderTime
			return c.Send(MessageChooseNewTimeOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		case StateUpdatingReminderTime:
			newTime := c.Text()
			_, err := time.Parse("15:04", newTime)
			if err != nil {
				return c.Send(ErrorInvalidTime, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTime)

			dueDateTime, err := time.ParseInLocation("2006-01-02 15:04", userState.DateTime, time.Local)
			if err != nil {
				return c.Send(ErrorInvalidDateTimeCompilation, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
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
				return c.Send(ErrorUpdateReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}

			// Запланируйте обновленное напоминание
			b.scheduleReminder(reminder)

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf(MessageUpdateReminderSuccess, reminder.Title), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})

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
			return c.Send(ErrorCallbackDataNull)
		}

		userState, ok := b.userStates[chatID]
		if !ok {
			return c.Send(StandardError, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}

		loc, err := time.LoadLocation("Europe/Moscow")
		if err != nil {
			log.Println("Error loading location:", err)
			return c.Send(ErrorLoadTimezone)
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
				return c.Send(MessageSetDateOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			default:
				return c.Send(ErrorCallbackDataNull, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}
			userState.DateTime = dateStr
			userState.State = StateSettingTime
			return c.Send(MessageChooseTimeOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2}, &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{{Text: KeyboardTime9, Data: "09:00"}, {Text: KeyboardTime12, Data: "12:00"}},
					{{Text: KeyboardTime15, Data: "15:00"}, {Text: KeyboardTime18, Data: "18:00"}},
					{{Text: KeyboardTime21, Data: "21:00"}},
				},
				RemoveKeyboard: true,
				ResizeKeyboard: true,
			})
		case StateSettingTime:
			newTimeStr := data
			_, err := time.Parse("15:04", newTimeStr)
			if err != nil {
				return c.Send(ErrorInvalidTime, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}
			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)

			dueDateTime, err := time.ParseInLocation("2006-01-02 15:04", userState.DateTime, loc)
			if err != nil {
				return c.Send(ErrorInvalidDateTimeCompilation, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}

			reminder := &models.Reminder{
				UserID:  uint(c.Sender().ID),
				Title:   userState.Title,
				DueDate: dueDateTime,
			}

			err = b.reminderService.CreateReminder(reminder)
			if err != nil {
				log.Println("Error creating reminder:", err)
				return c.Send(ErrorCreateReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}

			b.scheduleReminder(reminder)

			delete(b.userStates, chatID)
			return c.Send(fmt.Sprintf(MessageCreateReminderSuccess, reminder.Title), &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		case StateDeletingReminder:
			if strings.HasPrefix(data, "delete_") {
				userID := uint(c.Sender().ID)
				reminderIDStr := strings.TrimPrefix(data, "delete_")
				reminderID, err := strconv.Atoi(reminderIDStr)
				if err != nil {
					return c.Send(ErrorParseIdOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
				}

				err = b.reminderService.DeleteReminder(userID, uint(reminderID))
				if err != nil {
					return c.Send(ErrorFindReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
				}

				delete(b.userStates, chatID)
				return c.Send("♻♻♻")
			}
		case StateUpdatingReminderTitle:
			reminderIDStr := strings.TrimPrefix(data, "update_")
			reminderID, err := strconv.Atoi(reminderIDStr)
			if err != nil {
				return c.Send(ErrorParseIdOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
			}

			userState.State = StateUpdatingReminderTitle
			userState.ReminderID = reminderID

			return c.Send(MessageEnterNewNameOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		case StateViewingReminder:
			if strings.HasPrefix(data, "view_") {
				reminderIDStr := strings.TrimPrefix(data, "view_")
				reminderID, err := strconv.Atoi(reminderIDStr)
				if err != nil {
					return c.Send(ErrorParseIdOfReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
				}

				reminder, err := b.reminderService.GetReminderByID(uint(reminderID), userID)
				if err != nil {
					return c.Send(ErrorFindReminder, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
				}

				message := fmt.Sprintf(MessagePrintReminderData, reminder.Title, reminder.DueDate.In(time.Local).Format("2006-01-02 15:04"))
				return c.Send(message)
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
			return c.Send(ErrorGetUser, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}

		if len(user.Reminders) == 0 {
			return c.Send(MessageNoSuchReminders)
		}

		var buttons [][]telebot.InlineButton
		for _, reminder := range user.Reminders {
			buttons = append(buttons, []telebot.InlineButton{
				{Text: reminder.Title, Data: fmt.Sprintf("view_%d", reminder.ID)},
			})
		}

		// Устанавливаем состояние пользователя для просмотра напоминаний
		b.userStates[chatID] = &UserState{State: StateViewingReminder}

		return c.Send(MessageListReminders, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2}, &telebot.ReplyMarkup{
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
			return c.Send(ErrorGetUser, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2})
		}

		if len(user.Reminders) == 0 {
			return c.Send(MessageNoSuchReminders)
		}

		var buttons [][]telebot.InlineButton
		for _, reminder := range user.Reminders {
			buttons = append(buttons, []telebot.InlineButton{
				{Text: reminder.Title, Data: fmt.Sprintf("delete_%d", reminder.ID)},
			})
		}

		b.userStates[c.Chat().ID] = &UserState{State: StateDeletingReminder}

		return c.Send(MessageChooseReminderForDelete, &telebot.SendOptions{ParseMode: telebot.ModeMarkdownV2}, &telebot.ReplyMarkup{
			InlineKeyboard: buttons,
		})
	}
}

func (b *Bot) sendReminder(reminder *models.Reminder) {
	message := fmt.Sprintf(MessageReminderUser, reminder.Title, reminder.DueDate.In(time.Local).Format("2006-01-02 15:04"))
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
