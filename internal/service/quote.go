package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aluto/go-motivation/internal/entity"
	"github.com/aluto/go-motivation/internal/repository"
)

type QuoteService struct {
	quotes repository.QuoteRepository
	users  repository.UserRepository
}

func NewQuoteService(quotes repository.QuoteRepository, users repository.UserRepository) *QuoteService {
	return &QuoteService{quotes: quotes, users: users}
}

func (s *QuoteService) Add(ctx context.Context, q *entity.Quote) error {
	q.CreatedAt = time.Now()
	return s.quotes.Insert(ctx, q)
}

func (s *QuoteService) Count(ctx context.Context) (int64, error) {
	return s.quotes.Count(ctx)
}

func (s *QuoteService) GetNextForUser(ctx context.Context, chatID int64) (*entity.Quote, error) {
	total, err := s.quotes.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count quotes: %w", err)
	}
	if total == 0 {
		return nil, fmt.Errorf("no quotes in database")
	}

	user, err := s.users.GetByChatID(ctx, chatID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	pointer := user.QuotePointer % int(total)
	quote, err := s.quotes.GetByIndex(ctx, pointer)
	if err != nil {
		return nil, fmt.Errorf("get quote at index %d: %w", pointer, err)
	}

	if err := s.users.IncrementQuotePointer(ctx, chatID, total); err != nil {
		return nil, fmt.Errorf("increment pointer: %w", err)
	}

	return quote, nil
}

func (s *QuoteService) FormatQuote(q *entity.Quote) string {
	msg := "━━━━━━━━━━━━━━\n\n"
	msg += fmt.Sprintf("  _\"%s\"_\n", escapeMarkdownV2(q.Text))

	if q.Author != "" {
		msg += fmt.Sprintf("\n  *— %s*\n", escapeMarkdownV2(q.Author))
	}

	if q.Notes != "" {
		msg += fmt.Sprintf("\n  📝 %s\n", escapeMarkdownV2(q.Notes))
	}

	msg += "\n━━━━━━━━━━━━━━"
	return msg
}

func escapeMarkdownV2(s string) string {
	replacer := []string{
		"_", "\\_", "*", "\\*", "[", "\\[", "]", "\\]",
		"(", "\\(", ")", "\\)", "~", "\\~", "`", "\\`",
		">", "\\>", "#", "\\#", "+", "\\+", "-", "\\-",
		"=", "\\=", "|", "\\|", "{", "\\{", "}", "\\}",
		".", "\\.", "!", "\\!",
	}
	result := s
	for i := 0; i < len(replacer); i += 2 {
		result = replaceAll(result, replacer[i], replacer[i+1])
	}
	return result
}

func replaceAll(s, old, new string) string {
	var result []byte
	for i := 0; i < len(s); i++ {
		if string(s[i]) == old {
			result = append(result, []byte(new)...)
		} else {
			result = append(result, s[i])
		}
	}
	return string(result)
}
