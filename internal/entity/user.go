package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	StepAwaitingTimezone    = "awaiting_timezone"
	StepAwaitingQuotesCount = "awaiting_quotes_count"
	StepAwaitingWeekdays    = "awaiting_weekdays"
	StepAwaitingTimeHour    = "awaiting_time_%d_hour"
	StepAwaitingTimeMinute  = "awaiting_time_%d_minute"
	StepAwaitingEmailOptIn  = "awaiting_email_opt_in"
	StepAwaitingEmail       = "awaiting_email"
	StepCompleted           = "completed"
)

type SetupData struct {
	Timezone     string   `bson:"timezone,omitempty"`
	QuotesPerDay int      `bson:"quotes_per_day,omitempty"`
	Weekdays     []int    `bson:"weekdays,omitempty"`
	SendTimes    []string `bson:"send_times,omitempty"`
	CurrentHour  string   `bson:"current_hour,omitempty"`
}

type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty"    json:"id"`
	ChatID       int64         `bson:"chat_id"          json:"chat_id"`
	Timezone     string        `bson:"timezone"         json:"timezone"`
	QuotesPerDay int           `bson:"quotes_per_day"   json:"quotes_per_day"`
	Weekdays     []int         `bson:"weekdays"         json:"weekdays"`
	SendTimes    []string      `bson:"send_times"       json:"send_times"`
	QuotePointer int           `bson:"quote_pointer"    json:"quote_pointer"`
	Email        string        `bson:"email,omitempty"  json:"email,omitempty"`
	EmailEnabled bool          `bson:"email_enabled"    json:"email_enabled"`
	SetupStep    string        `bson:"setup_step"       json:"setup_step"`
	SetupData    *SetupData    `bson:"setup_data,omitempty" json:"setup_data,omitempty"`
	IsActive     bool          `bson:"is_active"        json:"is_active"`
	CreatedAt    time.Time     `bson:"created_at"       json:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"       json:"updated_at"`
}
