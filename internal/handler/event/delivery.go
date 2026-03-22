package event

import (
	"context"

	ev "github.com/aluto/go-motivation/internal/event"
	"github.com/aluto/go-motivation/internal/service"
	"github.com/aluto/go-motivation/internal/telegram"
	log "github.com/sirupsen/logrus"
)

type DeliveryHandler struct {
	quotes *service.QuoteService
	bot    *telegram.Bot
	bus    *ev.Bus
}

func NewDeliveryHandler(quotes *service.QuoteService, bot *telegram.Bot, bus *ev.Bus) *DeliveryHandler {
	return &DeliveryHandler{quotes: quotes, bot: bot, bus: bus}
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
	h.bus.Publish(ev.QuoteDelivered{
		ChatID:  req.ChatID,
		QuoteID: quote.ID.Hex(),
	})
}
