package bot

import (
	"emelyanenkoig/reminder/pkg/config"
	"emelyanenkoig/reminder/pkg/services"
	"gopkg.in/telebot.v3"
	"log"
	"sync"
	"time"
)

// Bot представляет структуру бота
type Bot struct {
	Bot             *telebot.Bot
	userService     *services.UserService
	reminderService *services.ReminderService
	userStates      map[int64]*UserState
	mu              sync.RWMutex // Mutex для синхронизации доступа к userStates
}

// NewBot создает новый экземпляр бота
func NewBot(userService *services.UserService, reminderService *services.ReminderService, config config.Config) (*Bot, error) {
	//botToken := "5621569001:AAF4zzjbRSON21P43bxgM95HLDGpr7WzZV8"

	b, err := telebot.NewBot(telebot.Settings{
		Token:  config.Bot.Token,
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
	b.Handle("/get", bot.HandleGetUser())
	b.Handle("/add", bot.HandleAddReminder())
	b.Handle("/list", bot.HandleListReminders())
	b.Handle("/delete", bot.HandleDeleteReminder()) // Добавили здесь
	b.Handle("/update", bot.HandleUpdateReminder()) // Добавляем обновление

	b.Handle(telebot.OnCallback, bot.HandleCallback())
	b.Handle(telebot.OnText, bot.HandleText())

	return bot, nil
}

// Start запускает бота
func (b *Bot) Start() {
	startMetricsServer()
	log.Println("Starting bot...")
	b.Bot.Start()
}
