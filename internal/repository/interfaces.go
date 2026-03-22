package repository

import (
	"context"

	"github.com/aluto/go-motivation/internal/entity"
)

type QuoteRepository interface {
	Insert(ctx context.Context, q *entity.Quote) error
	Count(ctx context.Context) (int64, error)
	GetByIndex(ctx context.Context, index int) (*entity.Quote, error)
}

type UserRepository interface {
	Upsert(ctx context.Context, u *entity.User) error
	GetByChatID(ctx context.Context, chatID int64) (*entity.User, error)
	GetAllActive(ctx context.Context) ([]entity.User, error)
	UpdateSetup(ctx context.Context, chatID int64, step string, data *entity.SetupData) error
	CompleteSetup(ctx context.Context, chatID int64, u *entity.User) error
	IncrementQuotePointer(ctx context.Context, chatID int64, totalQuotes int64) error
}
