package bot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"

	"github.com/aluto/go-motivation/internal/config"
	"github.com/aluto/go-motivation/internal/telegram"
)

type Router struct {
	bot   *telegram.Bot
	start *StartHandler
	setup *SetupHandler
	admin *AdminHandler
	cfg   *config.Config
}

func NewRouter(bot *telegram.Bot, start *StartHandler, setup *SetupHandler, admin *AdminHandler, cfg *config.Config) *Router {
	return &Router{bot: bot, start: start, setup: setup, admin: admin, cfg: cfg}
}

func (r *Router) Listen(ctx context.Context) {
	updates := r.bot.GetUpdatesChan()

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			if update.Message != nil {
				r.handleMessage(ctx, update.Message)
			}
			if update.CallbackQuery != nil {
				r.handleCallback(ctx, update.CallbackQuery)
			}
		}
	}
}

func (r *Router) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	isAdmin := chatID == r.cfg.AdminChatID
	log.Infof("message from chat_id=%d: %s", chatID, msg.Text)

	if msg.IsCommand() && msg.Command() == "start" {
		r.admin.CancelIfActive(chatID)
		r.start.Handle(ctx, chatID)
		return
	}

	switch msg.Text {
	case telegram.BtnSettings:
		r.start.HandleSettings(ctx, chatID, isAdmin)
		return
	case telegram.BtnReset:
		r.admin.CancelIfActive(chatID)
		r.start.Handle(ctx, chatID)
		return
	case telegram.BtnAddQuote:
		if isAdmin {
			r.admin.Handleadd(ctx, chatID)
		}
		return
	case telegram.BtnCount:
		if isAdmin {
			r.admin.HandleQuoteCount(ctx, chatID)
		}
		return
	}

	if isAdmin {
		if r.admin.HandleText(ctx, chatID, msg.Text) {
			return
		}
	}
}

func (r *Router) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	r.bot.AnswerCallback(cb.ID)
	chatID := cb.Message.Chat.ID
	messageID := cb.Message.MessageID
	data := cb.Data

	log.Infof("callback from chat_id=%d: %s", chatID, data)

	if data == "skip" && chatID == r.cfg.AdminChatID {
		if r.admin.HandleSkip(ctx, chatID) {
			return
		}
	}

	r.setup.HandleCallback(ctx, chatID, messageID, data)
}
