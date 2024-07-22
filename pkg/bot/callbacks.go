package bot

//import (
//	"fmt"
//	"gopkg.in/telebot.v3"
//	"time"
//)
//
//// Обработка выбора даты и времени
//func (b *Bot) HandleCallback() telebot.HandlerFunc {
//	return func(c telebot.Context) error {
//		chatID := c.Chat().ID
//		data := c.Callback().Data
//
//		userState, ok := b.userStates[chatID]
//		if ok && userState.State == StateCreatingRepositoryModel {
//			switch data {
//			case "create_YES_reminder":
//				return c.Send("Created a new reminder!")
//
//			case "create_NO_reminder":
//				delete(b.userStates, chatID)
//				return c.Send("Deleted new reminder!")
//			}
//		}
//
//		if ok && userState.State == StateCreatingDate {
//			switch data {
//			case "today":
//				today := time.Now().Format("2006-01-02")
//				userState.DateTime = today
//				userState.State = StateCreatingTime
//				return c.Send("Please enter the time (HH:MM).", &telebot.ReplyMarkup{
//					InlineKeyboard: [][]telebot.InlineButton{
//						{
//							{Text: "09:00", Data: "09:00"},
//							{Text: "12:00", Data: "12:00"},
//							{Text: "15:00", Data: "15:00"},
//							{Text: "18:00", Data: "18:00"},
//							{Text: "21:00", Data: "21:00"},
//							{Text: "set time", Data: "set_time"},
//						},
//					},
//					ResizeKeyboard: true,
//				})
//
//			case "tomorrow":
//				tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
//				userState.DateTime = tomorrow
//				userState.State = StateCreatingTime
//				return c.Send("Please enter the time (HH:MM).", &telebot.ReplyMarkup{
//					InlineKeyboard: [][]telebot.InlineButton{
//						{
//							{Text: "09:00", Data: "09:00"},
//							{Text: "12:00", Data: "12:00"},
//							{Text: "15:00", Data: "15:00"},
//							{Text: "18:00", Data: "18:00"},
//							{Text: "21:00", Data: "21:00"},
//							{Text: "set time", Data: "set_time"},
//						},
//					},
//					ResizeKeyboard: true,
//				})
//			case "set_date":
//				return c.Send("Please enter the due date (YYYY-MM-DD).")
//
//			default:
//				return c.Send("Unknown callback data. Please use /update_reminder to start updating a reminder.")
//
//			}
//
//		}
//
//		if ok && userState.State == StateCreatingTime {
//			var newTimeStr string
//			switch data {
//			case "09:00":
//				userState.State = StateCreatingRepositoryModel
//				newTimeStr = "09:00"
//			case "12:00":
//				userState.State = StateCreatingRepositoryModel
//				newTimeStr = "12:00"
//
//			case "15:00":
//				userState.State = StateCreatingRepositoryModel
//				newTimeStr = "15:00"
//			case "18:00":
//				userState.State = StateCreatingRepositoryModel
//				newTimeStr = "18:00"
//			case "21:00":
//				userState.State = StateCreatingRepositoryModel
//				newTimeStr = "21:00"
//			case "set_time":
//				userState.State = StateCreatingTime
//			default:
//				return c.Send("Unknown callback data.")
//
//			}
//			if userState.State != StateCreatingRepositoryModel {
//				return c.Send("Please enter the time (HH:MM).")
//			}
//
//			_, err := time.Parse("15:04", newTimeStr)
//			if err != nil {
//				return c.Send("Invalid time format. Please use HH:MM.")
//			}
//			userState.DateTime = fmt.Sprintf("%s %s", userState.DateTime, newTimeStr)
//			return c.Send(fmt.Sprintf("Reminder datetime has been created %v", userState.DateTime))
//
//		}
//		return nil
//	}
//}
