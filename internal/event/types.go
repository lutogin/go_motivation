package event

import "time"

type Event interface {
	EventName() string
}

type TickEvent struct {
	Time time.Time
}

func (e TickEvent) EventName() string { return "tick" }

type QuoteSendRequested struct {
	ChatID    int64
	Scheduled bool // true = sent by cron, false = triggered manually by user
}

func (e QuoteSendRequested) EventName() string { return "quote_send_requested" }

type QuoteDelivered struct {
	ChatID  int64
	QuoteID string
}

func (e QuoteDelivered) EventName() string { return "quote_delivered" }
