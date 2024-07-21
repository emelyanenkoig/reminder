package bot

import (
	"emelyanenkoig/reminder/pkg/services"
	"gopkg.in/telebot.v3"
	"log"
	"time"
)

// Bot представляет структуру бота
type Bot struct {
	Bot             *telebot.Bot
	userService     *services.UserService
	reminderService *services.ReminderService
	userStates      map[int64]*UserState
}

// NewBot создает новый экземпляр бота
func NewBot(userService *services.UserService, reminderService *services.ReminderService) (*Bot, error) {
	botToken := "5621569001:AAF4zzjbRSON21P43bxgM95HLDGpr7WzZV8"

	b, err := telebot.NewBot(telebot.Settings{
		Token:  botToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	bot := &Bot{
		Bot:             b,
		userService:     userService,
		reminderService: reminderService,
		userStates:      make(map[int64]*UserState),
	}

	b.Handle("/start", bot.HandleStart())
	b.Handle("/get_user", bot.HandleGetUser())
	b.Handle("/add_reminder", bot.HandleAddReminder())
	b.Handle("/get_reminders", bot.HandleGetReminders())
	b.Handle("/get_reminder", bot.HandleGetReminder())
	b.Handle("/update_reminder", bot.HandleUpdateReminder())
	b.Handle("/delete_reminder", bot.HandleDeleteReminder())

	b.Handle(telebot.OnCallback, bot.HandleCallback())
	b.Handle(telebot.OnText, bot.HandleText())
	b.Handle(telebot.OnText, bot.HandleDeleteText())
	b.Handle(telebot.OnText, bot.HandleUpdateText())

	return bot, nil
}

func (b *Bot) Start() {
	log.Println("Starting bot...")
	b.Bot.Start()
}
