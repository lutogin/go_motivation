package telegram

import (
	"errors"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

type Bot struct {
	api *tgbotapi.BotAPI
}

func NewBot(token string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}
	log.Infof("authorized on account %s", api.Self.UserName)
	return &Bot{api: api}, nil
}

func (b *Bot) API() *tgbotapi.BotAPI {
	return b.api
}

func (b *Bot) Send(chatID int64, text string, parseMode string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	if parseMode != "" {
		msg.ParseMode = parseMode
	}
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) SendWithReplyKeyboard(chatID int64, text string, keyboard tgbotapi.ReplyKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) SendWithInlineKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) EditMessageText(chatID int64, messageID int, text string, keyboard *tgbotapi.InlineKeyboardMarkup) error {
	edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
	if keyboard != nil {
		edit.ReplyMarkup = keyboard
	}
	_, err := b.api.Send(edit)
	return err
}

func (b *Bot) AnswerCallback(callbackID string) {
	callback := tgbotapi.NewCallback(callbackID, "")
	if _, err := b.api.Request(callback); err != nil {
		log.Warnf("answer callback: %v", err)
	}
}

func (b *Bot) DeleteMessage(chatID int64, messageID int) {
	if _, err := b.api.Request(tgbotapi.NewDeleteMessage(chatID, messageID)); err != nil {
		log.Warnf("delete message chat_id=%d msg_id=%d: %v", chatID, messageID, err)
	}
}

func (b *Bot) GetUpdatesChan() tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return b.api.GetUpdatesChan(u)
}

// IsBotBlocked reports whether the Telegram API error means the user blocked
// or deleted the bot. Telegram returns HTTP 403 for both cases.
func IsBotBlocked(err error) bool {
	if err == nil {
		return false
	}
	var tgErr *tgbotapi.Error
	if errors.As(err, &tgErr) {
		return tgErr.Code == 403
	}
	return false
}
