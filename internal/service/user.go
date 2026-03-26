package service

import (
	"context"
	"time"

	"github.com/aluto/go-motivation/internal/entity"
	"github.com/aluto/go-motivation/internal/repository"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetOrCreate(ctx context.Context, chatID int64) (*entity.User, error) {
	user, err := s.users.GetByChatID(ctx, chatID)
	if err == nil {
		return user, nil
	}

	newUser := &entity.User{
		ChatID:    chatID,
		SetupStep: entity.StepAwaitingTimezone,
		IsActive:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.users.Upsert(ctx, newUser); err != nil {
		return nil, err
	}
	return s.users.GetByChatID(ctx, chatID)
}

func (s *UserService) ResetSetup(ctx context.Context, chatID int64) error {
	return s.users.UpdateSetup(ctx, chatID, entity.StepAwaitingTimezone, &entity.SetupData{})
}

func (s *UserService) UpdateSetup(ctx context.Context, chatID int64, step string, data *entity.SetupData) error {
	return s.users.UpdateSetup(ctx, chatID, step, data)
}

func (s *UserService) CompleteSetup(ctx context.Context, chatID int64, data *entity.SetupData, emailAddr string, emailEnabled bool) error {
	user := &entity.User{
		Timezone:     data.Timezone,
		QuotesPerDay: data.QuotesPerDay,
		Weekdays:     data.Weekdays,
		SendTimes:    data.SendTimes,
		Email:        emailAddr,
		EmailEnabled: emailEnabled,
	}
	return s.users.CompleteSetup(ctx, chatID, user)
}

func (s *UserService) RestoreActive(ctx context.Context, chatID int64) error {
	return s.users.RestoreActive(ctx, chatID)
}

func (s *UserService) Deactivate(ctx context.Context, chatID int64) error {
	return s.users.Deactivate(ctx, chatID)
}

func (s *UserService) GetByChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	return s.users.GetByChatID(ctx, chatID)
}
