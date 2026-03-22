package service

import (
	"context"
	"fmt"
	"time"

	"github.com/aluto/go-motivation/internal/entity"
	"github.com/aluto/go-motivation/internal/repository"
	log "github.com/sirupsen/logrus"
)

type SchedulerService struct {
	users repository.UserRepository
}

func NewSchedulerService(users repository.UserRepository) *SchedulerService {
	return &SchedulerService{users: users}
}

func (s *SchedulerService) FindEligibleUsers(ctx context.Context, now time.Time) ([]entity.User, error) {
	allUsers, err := s.users.GetAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active users: %w", err)
	}

	var eligible []entity.User
	for _, u := range allUsers {
		loc, err := time.LoadLocation(u.Timezone)
		if err != nil {
			log.Warnf("invalid timezone %q for chat_id=%d, skipping", u.Timezone, u.ChatID)
			continue
		}

		localTime := now.In(loc)
		localWeekday := int(localTime.Weekday())
		localHHMM := localTime.Format("15:04")

		if !containsInt(u.Weekdays, localWeekday) {
			continue
		}

		for _, t := range u.SendTimes {
			if t == localHHMM {
				eligible = append(eligible, u)
				break
			}
		}
	}

	return eligible, nil
}

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
