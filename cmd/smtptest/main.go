// Local SMTP test: run from repository root
//
//	go run ./cmd/smtptest
//
// Required in .env or environment: SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM
// (same as the main bot). Sends the same HTML as production (internal/email.QuoteEmailRFC822).
package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"

	"github.com/aluto/go-motivation/internal/email"
	"github.com/aluto/go-motivation/internal/entity"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Hardcoded recipient for local runs; replace with your inbox.
const testTo = "lutogin.v8@gmail.com"

type smtpEnv struct {
	Host string `env:"SMTP_HOST" env-default:"smtp.gmail.com"`
	Port int    `env:"SMTP_PORT" env-default:"587"`
	User string `env:"SMTP_USER"`
	Pass string `env:"SMTP_PASS"`
	From string `env:"SMTP_FROM"`
}

func main() {
	_ = godotenv.Load()

	var cfg smtpEnv
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("read env: %v", err)
	}
	if cfg.User == "" || cfg.Pass == "" || cfg.From == "" {
		log.Fatal("set SMTP_USER, SMTP_PASS, and SMTP_FROM in .env or environment")
	}

	mock := &entity.Quote{
		Text:   "The only way to do great work is to love what you do.",
		Author: "Steve Jobs",
	}

	raw := email.QuoteEmailRFC822(cfg.From, testTo, mock)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	log.Printf("sending HTML quote to %s via %s (subject: %q) ...", testTo, addr, email.QuoteEmailSubject)
	if err := smtp.SendMail(addr, auth, cfg.From, []string{testTo}, raw); err != nil {
		log.Fatalf("SendMail: %v", err)
	}

	log.Println("ok: email sent")
	os.Exit(0)
}
