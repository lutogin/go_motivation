package bot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aluto/go-motivation/internal/config"
	"github.com/aluto/go-motivation/internal/email"
	"github.com/aluto/go-motivation/internal/entity"
	"github.com/aluto/go-motivation/internal/service"
	"github.com/aluto/go-motivation/internal/telegram"
	log "github.com/sirupsen/logrus"
)

type SetupHandler struct {
	bot   *telegram.Bot
	users *service.UserService
	cfg   *config.Config
}

func NewSetupHandler(bot *telegram.Bot, users *service.UserService, cfg *config.Config) *SetupHandler {
	return &SetupHandler{bot: bot, users: users, cfg: cfg}
}

func (h *SetupHandler) HandleCallback(ctx context.Context, chatID int64, messageID int, data string) {
	user, err := h.users.GetByChatID(ctx, chatID)
	if err != nil {
		log.Errorf("get user for setup: %v", err)
		return
	}

	switch {
	case strings.HasPrefix(data, "tz_region:"):
		h.handleTimezoneRegion(ctx, chatID, messageID, data, user)
	case strings.HasPrefix(data, "tz:"):
		h.handleTimezoneSelect(ctx, chatID, messageID, data, user)
	case strings.HasPrefix(data, "count:"):
		h.handleQuotesCount(ctx, chatID, messageID, data, user)
	case strings.HasPrefix(data, "day:"):
		h.handleWeekdayToggle(ctx, chatID, messageID, data, user)
	case data == "days_done":
		h.handleWeekdaysDone(ctx, chatID, messageID, user)
	case strings.HasPrefix(data, "hour:"):
		h.handleHourSelect(ctx, chatID, messageID, data, user)
	case strings.HasPrefix(data, "minute:"):
		h.handleMinuteSelect(ctx, chatID, messageID, data, user)
	case data == "email_opt_in:yes":
		h.handleEmailOptInYes(ctx, chatID, messageID, user)
	case data == "email_opt_in:no":
		h.handleEmailOptInNo(ctx, chatID, messageID, user)
	}
}

func (h *SetupHandler) HandleEmailInput(ctx context.Context, chatID int64, text string) bool {
	user, err := h.users.GetByChatID(ctx, chatID)
	if err != nil || user.SetupStep != entity.StepAwaitingEmail {
		return false
	}

	text = strings.TrimSpace(text)
	if !email.ValidateEmail(text) {
		h.bot.Send(chatID, "❌ Неверный формат email. Попробуй ещё раз:", "")
		return true
	}

	h.completeSetup(ctx, chatID, 0, user.SetupData, text, true)
	return true
}

func (h *SetupHandler) handleTimezoneRegion(ctx context.Context, chatID int64, messageID int, data string, user *entity.User) {
	region := strings.TrimPrefix(data, "tz_region:")

	if region == "back" {
		kb := telegram.TimezoneRegionsKeyboard()
		h.bot.EditMessageText(chatID, messageID, "🌍 Выбери свой регион:", &kb)
		return
	}

	kb := telegram.TimezoneCitiesKeyboard(region)
	h.bot.EditMessageText(chatID, messageID, fmt.Sprintf("🌍 Выбери таймзону (%s):", region), &kb)
}

func (h *SetupHandler) handleTimezoneSelect(ctx context.Context, chatID int64, messageID int, data string, user *entity.User) {
	if user.SetupStep != entity.StepAwaitingTimezone {
		return
	}

	tz := strings.TrimPrefix(data, "tz:")
	setupData := &entity.SetupData{Timezone: tz}

	if err := h.users.UpdateSetup(ctx, chatID, entity.StepAwaitingQuotesCount, setupData); err != nil {
		log.Errorf("update timezone: %v", err)
		return
	}

	kb := telegram.QuotesCountKeyboard()
	h.bot.EditMessageText(chatID, messageID,
		fmt.Sprintf("✅ Таймзона: %s\n\nСколько цитат в день ты хочешь получать?", tz), &kb)
}

func (h *SetupHandler) handleQuotesCount(ctx context.Context, chatID int64, messageID int, data string, user *entity.User) {
	if user.SetupStep != entity.StepAwaitingQuotesCount {
		return
	}

	countStr := strings.TrimPrefix(data, "count:")
	count, _ := strconv.Atoi(countStr)
	if count < 1 || count > 3 {
		return
	}

	setupData := user.SetupData
	if setupData == nil {
		setupData = &entity.SetupData{}
	}
	setupData.QuotesPerDay = count

	if err := h.users.UpdateSetup(ctx, chatID, entity.StepAwaitingWeekdays, setupData); err != nil {
		log.Errorf("update quotes count: %v", err)
		return
	}

	kb := telegram.WeekdaysKeyboard(nil)
	h.bot.EditMessageText(chatID, messageID,
		fmt.Sprintf("✅ Цитат в день: %d\n\nВыбери дни недели для рассылки:", count), &kb)
}

func (h *SetupHandler) handleWeekdayToggle(ctx context.Context, chatID int64, messageID int, data string, user *entity.User) {
	if user.SetupStep != entity.StepAwaitingWeekdays {
		return
	}

	dayStr := strings.TrimPrefix(data, "day:")
	day, _ := strconv.Atoi(dayStr)

	setupData := user.SetupData
	if setupData == nil {
		setupData = &entity.SetupData{}
	}

	found := false
	var newDays []int
	for _, d := range setupData.Weekdays {
		if d == day {
			found = true
			continue
		}
		newDays = append(newDays, d)
	}
	if !found {
		newDays = append(newDays, day)
	}
	setupData.Weekdays = newDays

	if err := h.users.UpdateSetup(ctx, chatID, entity.StepAwaitingWeekdays, setupData); err != nil {
		log.Errorf("update weekdays: %v", err)
		return
	}

	kb := telegram.WeekdaysKeyboard(setupData.Weekdays)
	h.bot.EditMessageText(chatID, messageID, "📅 Выбери дни недели для рассылки:", &kb)
}

func (h *SetupHandler) handleWeekdaysDone(ctx context.Context, chatID int64, messageID int, user *entity.User) {
	if user.SetupStep != entity.StepAwaitingWeekdays {
		return
	}

	setupData := user.SetupData
	if setupData == nil || len(setupData.Weekdays) == 0 {
		h.bot.Send(chatID, "⚠️ Выбери хотя бы один день!", "")
		return
	}

	step := fmt.Sprintf(entity.StepAwaitingTimeHour, 1)
	if err := h.users.UpdateSetup(ctx, chatID, step, setupData); err != nil {
		log.Errorf("update to time selection: %v", err)
		return
	}

	kb := telegram.HourKeyboard()
	h.bot.EditMessageText(chatID, messageID,
		"🕐 Выбери час для цитаты #1:", &kb)
}

func (h *SetupHandler) handleHourSelect(ctx context.Context, chatID int64, messageID int, data string, user *entity.User) {
	hour := strings.TrimPrefix(data, "hour:")

	setupData := user.SetupData
	if setupData == nil {
		setupData = &entity.SetupData{}
	}
	setupData.CurrentHour = hour

	currentStep := user.SetupStep
	n := extractTimeSlotNumber(currentStep)
	if n == 0 {
		return
	}

	minuteStep := fmt.Sprintf(entity.StepAwaitingTimeMinute, n)
	if err := h.users.UpdateSetup(ctx, chatID, minuteStep, setupData); err != nil {
		log.Errorf("update hour: %v", err)
		return
	}

	kb := telegram.MinuteKeyboard()
	h.bot.EditMessageText(chatID, messageID,
		fmt.Sprintf("🕐 Выбери минуты для цитаты #%d (час: %s):", n, hour), &kb)
}

func (h *SetupHandler) handleMinuteSelect(ctx context.Context, chatID int64, messageID int, data string, user *entity.User) {
	minute := strings.TrimPrefix(data, "minute:")

	setupData := user.SetupData
	if setupData == nil {
		setupData = &entity.SetupData{}
	}

	currentStep := user.SetupStep
	n := extractTimeSlotNumber(currentStep)
	if n == 0 {
		return
	}

	timeStr := setupData.CurrentHour + ":" + minute
	setupData.SendTimes = append(setupData.SendTimes, timeStr)
	setupData.CurrentHour = ""

	if n < setupData.QuotesPerDay {
		nextStep := fmt.Sprintf(entity.StepAwaitingTimeHour, n+1)
		if err := h.users.UpdateSetup(ctx, chatID, nextStep, setupData); err != nil {
			log.Errorf("update to next time slot: %v", err)
			return
		}

		kb := telegram.HourKeyboard()
		h.bot.EditMessageText(chatID, messageID,
			fmt.Sprintf("✅ Цитата #%d в %s\n\n🕐 Выбери час для цитаты #%d:", n, timeStr, n+1), &kb)
		return
	}

	if err := h.users.UpdateSetup(ctx, chatID, entity.StepAwaitingEmailOptIn, setupData); err != nil {
		log.Errorf("update to email opt-in: %v", err)
		return
	}

	kb := telegram.EmailOptInKeyboard()
	h.bot.EditMessageText(chatID, messageID,
		fmt.Sprintf("✅ Цитата #%d в %s\n\n📧 Хочешь также получать цитаты на email?", n, timeStr), &kb)
}

func (h *SetupHandler) handleEmailOptInYes(ctx context.Context, chatID int64, messageID int, user *entity.User) {
	if user.SetupStep != entity.StepAwaitingEmailOptIn {
		return
	}

	if err := h.users.UpdateSetup(ctx, chatID, entity.StepAwaitingEmail, user.SetupData); err != nil {
		log.Errorf("update to email input: %v", err)
		return
	}

	h.bot.EditMessageText(chatID, messageID, "📧 Введи свой email:", nil)
}

func (h *SetupHandler) handleEmailOptInNo(ctx context.Context, chatID int64, messageID int, user *entity.User) {
	if user.SetupStep != entity.StepAwaitingEmailOptIn {
		return
	}

	h.completeSetup(ctx, chatID, messageID, user.SetupData, "", false)
}

func (h *SetupHandler) completeSetup(ctx context.Context, chatID int64, messageID int, setupData *entity.SetupData, emailAddr string, emailEnabled bool) {
	if err := h.users.CompleteSetup(ctx, chatID, setupData, emailAddr, emailEnabled); err != nil {
		log.Errorf("complete setup: %v", err)
		return
	}

	dayNames := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
	days := ""
	for i, d := range setupData.Weekdays {
		if i > 0 {
			days += ", "
		}
		days += dayNames[d]
	}

	times := ""
	for i, t := range setupData.SendTimes {
		if i > 0 {
			times += ", "
		}
		times += t
	}

	summary := fmt.Sprintf("🎉 Настройка завершена!\n\n"+
		"🌍 Таймзона: %s\n"+
		"📊 Цитат в день: %d\n"+
		"📅 Дни: %s\n"+
		"🕐 Время: %s",
		setupData.Timezone, setupData.QuotesPerDay, days, times)

	if emailEnabled {
		summary += fmt.Sprintf("\n📧 Email: %s", emailAddr)
	}

	summary += "\n\nОжидай свою первую цитату! ✨"

	if messageID > 0 {
		h.bot.EditMessageText(chatID, messageID, summary, nil)
	} else {
		h.bot.Send(chatID, summary, "")
	}

	isAdmin := chatID == h.cfg.AdminChatID
	kb := telegram.MainMenuKeyboard(isAdmin)
	h.bot.SendWithReplyKeyboard(chatID, "Используй кнопки ниже для управления 👇", kb)
}

func extractTimeSlotNumber(step string) int {
	var n int
	if _, err := fmt.Sscanf(step, "awaiting_time_%d_hour", &n); err == nil {
		return n
	}
	if _, err := fmt.Sscanf(step, "awaiting_time_%d_minute", &n); err == nil {
		return n
	}
	return 0
}
