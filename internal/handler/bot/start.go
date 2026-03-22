package bot

import (
	"context"
	"fmt"

	"github.com/aluto/go-motivation/internal/service"
	"github.com/aluto/go-motivation/internal/telegram"
	log "github.com/sirupsen/logrus"
)

type StartHandler struct {
	bot   *telegram.Bot
	users *service.UserService
}

func NewStartHandler(bot *telegram.Bot, users *service.UserService) *StartHandler {
	return &StartHandler{bot: bot, users: users}
}

func (h *StartHandler) Handle(ctx context.Context, chatID int64) {
	if _, err := h.users.GetOrCreate(ctx, chatID); err != nil {
		log.Errorf("get or create user: %v", err)
		return
	}

	if err := h.users.ResetSetup(ctx, chatID); err != nil {
		log.Errorf("reset setup: %v", err)
		return
	}

	kb := telegram.TimezoneRegionsKeyboard()
	if err := h.bot.SendWithInlineKeyboard(chatID,
		"👋 Добро пожаловать! Я буду отправлять тебе мотивационные цитаты.\n\n"+
			"Для начала выбери свой регион:", kb); err != nil {
		log.Errorf("send start: %v", err)
	}
}

func (h *StartHandler) HandleSettings(ctx context.Context, chatID int64) {
	user, err := h.users.GetByChatID(ctx, chatID)
	if err != nil {
		log.Errorf("get user: %v", err)
		return
	}

	days := ""
	dayNames := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
	for i, d := range user.Weekdays {
		if i > 0 {
			days += ", "
		}
		days += dayNames[d]
	}

	times := ""
	for i, t := range user.SendTimes {
		if i > 0 {
			times += ", "
		}
		times += t
	}

	text := fmt.Sprintf("⚙️ Текущие настройки:\n\n"+
		"🌍 Таймзона: %s\n"+
		"📊 Цитат в день: %d\n"+
		"📅 Дни: %s\n"+
		"🕐 Время: %s\n\n"+
		"Чтобы изменить — /start",
		user.Timezone, user.QuotesPerDay, days, times)

	if err := h.bot.Send(chatID, text, ""); err != nil {
		log.Errorf("send settings: %v", err)
	}
}
