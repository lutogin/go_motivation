package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aluto/go-motivation/internal/entity"
	"github.com/aluto/go-motivation/internal/service"
	"github.com/aluto/go-motivation/internal/telegram"
	log "github.com/sirupsen/logrus"
)

type adminState struct {
	step        string
	quote       entity.Quote
	promptMsgID int // message ID of the last bot prompt, so we can delete it
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
	promptID := h.bot.SendTracked(chatID, "📝 Введи текст цитаты:")

	h.mu.Lock()
	h.states[chatID] = &adminState{step: "text", promptMsgID: promptID}
	h.mu.Unlock()
}

func (h *AdminHandler) HandleQuoteCount(ctx context.Context, chatID int64) {
	count, err := h.quotes.Count(ctx)
	if err != nil {
		log.Errorf("count quotes: %v", err)
		return
	}
	msgID := h.bot.SendTracked(chatID, fmt.Sprintf("📊 Всего цитат в базе: %d", count))
	if msgID != 0 {
		h.bot.DeleteMessageAfter(chatID, msgID, 5*time.Second)
	}
}

func (h *AdminHandler) HandleText(ctx context.Context, chatID int64, userMsgID int, text string) bool {
	h.mu.Lock()
	state, ok := h.states[chatID]
	h.mu.Unlock()
	if !ok {
		return false
	}

	// Delete the previous bot prompt and the user's input message.
	h.bot.DeleteMessage(chatID, state.promptMsgID)
	h.bot.DeleteMessage(chatID, userMsgID)

	switch state.step {
	case "text":
		state.quote.Text = text
		state.step = "author"
		kb := telegram.SkipKeyboard()
		state.promptMsgID = h.bot.SendWithInlineKeyboardTracked(chatID, "✍️ Введи автора цитаты (или пропусти):", kb)

	case "author":
		state.quote.Author = text
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
		h.saveQuote(ctx, chatID, state)
	default:
		return false
	}
	return true
}

func (h *AdminHandler) saveQuote(ctx context.Context, chatID int64, state *adminState) {
	// Delete the last prompt that is still visible in chat.
	h.bot.DeleteMessage(chatID, state.promptMsgID)

	if err := h.quotes.Add(ctx, &state.quote); err != nil {
		log.Errorf("save quote: %v", err)
		h.bot.Send(chatID, "❌ Ошибка при сохранении цитаты", "")
		return
	}

	summary := fmt.Sprintf("✅ Цитата добавлена!\n\n📖 %s", state.quote.Text)
	if state.quote.Author != "" {
		summary += fmt.Sprintf("\n✍️ %s", state.quote.Author)
	}

	msgID := h.bot.SendTracked(chatID, summary)
	if msgID != 0 {
		h.bot.DeleteMessageAfter(chatID, msgID, 2*time.Second)
	}

	h.mu.Lock()
	delete(h.states, chatID)
	h.mu.Unlock()
}
