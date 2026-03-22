package bot

import (
	"context"
	"fmt"
	"sync"

	"github.com/aluto/go-motivation/internal/entity"
	"github.com/aluto/go-motivation/internal/service"
	"github.com/aluto/go-motivation/internal/telegram"
	log "github.com/sirupsen/logrus"
)

type adminState struct {
	step  string
	quote entity.Quote
}

type AdminHandler struct {
	bot    *telegram.Bot
	quotes *service.QuoteService
	mu     sync.Mutex
	states map[int64]*adminState
}

func NewAdminHandler(bot *telegram.Bot, quotes *service.QuoteService) *AdminHandler {
	return &AdminHandler{
		bot:    bot,
		quotes: quotes,
		states: make(map[int64]*adminState),
	}
}

func (h *AdminHandler) CancelIfActive(chatID int64) {
	h.mu.Lock()
	delete(h.states, chatID)
	h.mu.Unlock()
}

func (h *AdminHandler) Handleadd(ctx context.Context, chatID int64) {
	h.mu.Lock()
	h.states[chatID] = &adminState{step: "text"}
	h.mu.Unlock()

	h.bot.Send(chatID, "📝 Введи текст цитаты:", "")
}

func (h *AdminHandler) HandleQuoteCount(ctx context.Context, chatID int64) {
	count, err := h.quotes.Count(ctx)
	if err != nil {
		log.Errorf("count quotes: %v", err)
		return
	}
	h.bot.Send(chatID, fmt.Sprintf("📊 Всего цитат в базе: %d", count), "")
}

func (h *AdminHandler) HandleText(ctx context.Context, chatID int64, text string) bool {
	h.mu.Lock()
	state, ok := h.states[chatID]
	h.mu.Unlock()
	if !ok {
		return false
	}

	switch state.step {
	case "text":
		state.quote.Text = text
		state.step = "author"
		kb := telegram.SkipKeyboard()
		h.bot.SendWithInlineKeyboard(chatID, "✍️ Введи автора цитаты (или пропусти):", kb)

	case "author":
		state.quote.Author = text
		state.step = "notes"
		kb := telegram.SkipKeyboard()
		h.bot.SendWithInlineKeyboard(chatID, "📝 Введи примечания (или пропусти):", kb)

	case "notes":
		state.quote.Notes = text
		state.step = "category"
		kb := telegram.SkipKeyboard()
		h.bot.SendWithInlineKeyboard(chatID, "🏷 Введи категорию (или пропусти):", kb)

	case "category":
		state.quote.Category = text
		h.saveQuote(ctx, chatID, state)

	default:
		return false
	}

	return true
}

func (h *AdminHandler) HandleSkip(ctx context.Context, chatID int64) bool {
	h.mu.Lock()
	state, ok := h.states[chatID]
	h.mu.Unlock()
	if !ok {
		return false
	}

	switch state.step {
	case "author":
		state.step = "notes"
		kb := telegram.SkipKeyboard()
		h.bot.SendWithInlineKeyboard(chatID, "📝 Введи примечания (или пропусти):", kb)
	case "notes":
		state.step = "category"
		kb := telegram.SkipKeyboard()
		h.bot.SendWithInlineKeyboard(chatID, "🏷 Введи категорию (или пропусти):", kb)
	case "category":
		h.saveQuote(ctx, chatID, state)
	default:
		return false
	}
	return true
}

func (h *AdminHandler) saveQuote(ctx context.Context, chatID int64, state *adminState) {
	if err := h.quotes.Add(ctx, &state.quote); err != nil {
		log.Errorf("save quote: %v", err)
		h.bot.Send(chatID, "❌ Ошибка при сохранении цитаты", "")
		return
	}

	summary := fmt.Sprintf("✅ Цитата добавлена!\n\n"+
		"📖 %s", state.quote.Text)
	if state.quote.Author != "" {
		summary += fmt.Sprintf("\n✍️ %s", state.quote.Author)
	}
	if state.quote.Notes != "" {
		summary += fmt.Sprintf("\n📝 %s", state.quote.Notes)
	}
	if state.quote.Category != "" {
		summary += fmt.Sprintf("\n🏷 %s", state.quote.Category)
	}

	h.bot.Send(chatID, summary, "")

	h.mu.Lock()
	delete(h.states, chatID)
	h.mu.Unlock()
}
