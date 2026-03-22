package scheduler

import (
	"time"

	"github.com/aluto/go-motivation/internal/event"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

type Cron struct {
	c   *cron.Cron
	bus *event.Bus
}

func NewCron(bus *event.Bus) *Cron {
	return &Cron{
		c:   cron.New(),
		bus: bus,
	}
}

func (s *Cron) Start() {
	s.c.AddFunc("*/5 * * * *", func() {
		now := time.Now().UTC()
		log.Debugf("cron tick at %s", now.Format(time.RFC3339))
		s.bus.Publish(event.TickEvent{Time: now})
	})
	s.c.Start()
	log.Info("cron scheduler started (every 5 minutes)")
}

func (s *Cron) Stop() {
	s.c.Stop()
}
