package event

import (
	"context"

	ev "github.com/aluto/go-motivation/internal/event"
	"github.com/aluto/go-motivation/internal/service"
	log "github.com/sirupsen/logrus"
)

type SchedulerHandler struct {
	scheduler *service.SchedulerService
	bus       *ev.Bus
}

func NewSchedulerHandler(scheduler *service.SchedulerService, bus *ev.Bus) *SchedulerHandler {
	return &SchedulerHandler{scheduler: scheduler, bus: bus}
}

func (h *SchedulerHandler) Register() {
	h.bus.Subscribe("tick", h.handle)
}

func (h *SchedulerHandler) handle(ctx context.Context, e ev.Event) {
	tick, ok := e.(ev.TickEvent)
	if !ok {
		return
	}

	users, err := h.scheduler.FindEligibleUsers(ctx, tick.Time)
	if err != nil {
		log.Errorf("find eligible users: %v", err)
		return
	}

	if len(users) > 0 {
		log.Infof("found %d eligible users for quote delivery", len(users))
	}

	for _, u := range users {
		h.bus.Publish(ev.QuoteSendRequested{ChatID: u.ChatID, Scheduled: true})
	}
}
