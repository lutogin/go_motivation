package event

import (
	"context"

	"github.com/aluto/go-motivation/internal/email"
	ev "github.com/aluto/go-motivation/internal/event"
	"github.com/aluto/go-motivation/internal/service"
	"github.com/aluto/go-motivation/internal/telegram"
	log "github.com/sirupsen/logrus"
)

type DeliveryHandler struct {
	quotes *service.QuoteService
	users  *service.UserService
	bot    *telegram.Bot
	bus    *ev.Bus
	mail   *email.Sender
}

func NewDeliveryHandler(quotes *service.QuoteService, users *service.UserService, bot *telegram.Bot, bus *ev.Bus, mail *email.Sender) *DeliveryHandler {
	return &DeliveryHandler{quotes: quotes, users: users, bot: bot, bus: bus, mail: mail}
}

func (h *DeliveryHandler) Register() {
	h.bus.Subscribe("quote_send_requested", h.handle)
}

func (h *DeliveryHandler) handle(ctx context.Context, e ev.Event) {
	req, ok := e.(ev.QuoteSendRequested)
	if !ok {
		return
	}

	quote, err := h.quotes.GetNextForUser(ctx, req.ChatID)
	if err != nil {
		log.Errorf("get next quote for chat_id=%d: %v", req.ChatID, err)
		return
	}

	formatted := h.quotes.FormatQuote(quote)
	if err := h.bot.Send(req.ChatID, formatted, "MarkdownV2"); err != nil {
		log.Errorf("send quote to chat_id=%d: %v", req.ChatID, err)
		return
	}

	log.Infof("delivered quote %s to chat_id=%d", quote.ID.Hex(), req.ChatID)

	user, err := h.users.GetByChatID(ctx, req.ChatID)
	if err == nil && user.EmailEnabled && user.Email != "" {
		if err := h.mail.SendQuote(user.Email, quote); err != nil {
			log.Errorf("send email quote to %s: %v", user.Email, err)
		}
	}

	h.bus.Publish(ev.QuoteDelivered{
		ChatID:  req.ChatID,
		QuoteID: quote.ID.Hex(),
	})
}
