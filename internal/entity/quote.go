package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Quote struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Text      string        `bson:"text"          json:"text"`
	Author    string        `bson:"author,omitempty" json:"author,omitempty"`
	Notes     string        `bson:"notes,omitempty"  json:"notes,omitempty"`
	Category  string        `bson:"category,omitempty" json:"category,omitempty"`
	CreatedAt time.Time     `bson:"created_at"    json:"created_at"`
}
