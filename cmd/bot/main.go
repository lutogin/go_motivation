package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/dig"

	"github.com/aluto/go-motivation/internal/config"
	"github.com/aluto/go-motivation/internal/email"
	"github.com/aluto/go-motivation/internal/event"
	bothandler "github.com/aluto/go-motivation/internal/handler/bot"
	eventhandler "github.com/aluto/go-motivation/internal/handler/event"
	"github.com/aluto/go-motivation/internal/repository"
	mongorepo "github.com/aluto/go-motivation/internal/repository/mongo"
	"github.com/aluto/go-motivation/internal/scheduler"
	"github.com/aluto/go-motivation/internal/service"
	"github.com/aluto/go-motivation/internal/telegram"
)

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)

	c := dig.New()

	provideAll(c)

	if err := c.Invoke(run); err != nil {
		log.Fatalf("failed to start: %v", err)
	}
}

func provideAll(c *dig.Container) {
	c.Provide(config.Load)

	c.Provide(func(cfg *config.Config) (*mongo.Database, error) {
		client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
		if err != nil {
			return nil, err
		}
		return client.Database(cfg.MongoDB), nil
	})

	c.Provide(func(cfg *config.Config) (*telegram.Bot, error) {
		return telegram.NewBot(cfg.BotToken)
	})

	c.Provide(func() *event.Bus {
		return event.NewBus(256)
	})

	c.Provide(func(db *mongo.Database) repository.QuoteRepository {
		return mongorepo.NewQuoteRepo(db)
	})
	c.Provide(func(db *mongo.Database) repository.UserRepository {
		return mongorepo.NewUserRepo(db)
	})

	c.Provide(email.NewSender)

	c.Provide(service.NewQuoteService)
	c.Provide(service.NewUserService)
	c.Provide(service.NewSchedulerService)

	c.Provide(bothandler.NewStartHandler)
	c.Provide(bothandler.NewSetupHandler)
	c.Provide(bothandler.NewAdminHandler)
	c.Provide(bothandler.NewRouter)

	c.Provide(eventhandler.NewSchedulerHandler)
	c.Provide(eventhandler.NewDeliveryHandler)

	c.Provide(scheduler.NewCron)
}

type appDeps struct {
	dig.In

	Router    *bothandler.Router
	Bus       *event.Bus
	Cron      *scheduler.Cron
	SchedH    *eventhandler.SchedulerHandler
	DeliveryH *eventhandler.DeliveryHandler
}

func run(deps appDeps) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deps.SchedH.Register()
	deps.DeliveryH.Register()

	deps.Bus.Start(ctx)
	log.Info("event bus started")

	deps.Cron.Start()
	defer deps.Cron.Stop()

	go deps.Router.Listen(ctx)
	log.Info("bot is listening for updates")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	cancel()
}
