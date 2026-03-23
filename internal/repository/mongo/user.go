package mongo

import (
	"context"
	"time"

	"github.com/aluto/go-motivation/internal/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepo struct {
	col *mongo.Collection
}

func NewUserRepo(db *mongo.Database) *UserRepo {
	col := db.Collection("users")
	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "chat_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &UserRepo{col: col}
}

func (r *UserRepo) Upsert(ctx context.Context, u *entity.User) error {
	now := time.Now()

	filter := bson.M{"chat_id": u.ChatID}
	update := bson.M{
		"$set": bson.M{
			"chat_id":        u.ChatID,
			"timezone":       u.Timezone,
			"quotes_per_day": u.QuotesPerDay,
			"weekdays":       u.Weekdays,
			"send_times":     u.SendTimes,
			"quote_pointer":  u.QuotePointer,
			"email":          u.Email,
			"email_enabled":  u.EmailEnabled,
			"setup_step":     u.SetupStep,
			"setup_data":     u.SetupData,
			"is_active":      u.IsActive,
			"updated_at":     now,
		},
		"$setOnInsert": bson.M{
			"created_at": now,
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.col.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *UserRepo) GetByChatID(ctx context.Context, chatID int64) (*entity.User, error) {
	var u entity.User
	err := r.col.FindOne(ctx, bson.M{"chat_id": chatID}).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetAllActive(ctx context.Context) ([]entity.User, error) {
	filter := bson.M{
		"is_active":  true,
		"setup_step": entity.StepCompleted,
	}
	cursor, err := r.col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepo) UpdateSetup(ctx context.Context, chatID int64, step string, data *entity.SetupData) error {
	update := bson.M{
		"$set": bson.M{
			"setup_step": step,
			"setup_data": data,
			"updated_at": time.Now(),
		},
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"chat_id": chatID}, update)
	return err
}

func (r *UserRepo) CompleteSetup(ctx context.Context, chatID int64, u *entity.User) error {
	update := bson.M{
		"$set": bson.M{
			"timezone":       u.Timezone,
			"quotes_per_day": u.QuotesPerDay,
			"weekdays":       u.Weekdays,
			"send_times":     u.SendTimes,
			"email":          u.Email,
			"email_enabled":  u.EmailEnabled,
			"setup_step":     entity.StepCompleted,
			"setup_data":     nil,
			"is_active":      true,
			"updated_at":     time.Now(),
		},
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"chat_id": chatID}, update)
	return err
}

func (r *UserRepo) RestoreActive(ctx context.Context, chatID int64) error {
	update := bson.M{
		"$set": bson.M{
			"setup_step": entity.StepCompleted,
			"setup_data": nil,
			"is_active":  true,
			"updated_at": time.Now(),
		},
	}
	_, err := r.col.UpdateOne(ctx, bson.M{"chat_id": chatID}, update)
	return err
}

func (r *UserRepo) IncrementQuotePointer(ctx context.Context, chatID int64, totalQuotes int64) error {
	_, err := r.col.UpdateOne(ctx, bson.M{"chat_id": chatID}, bson.A{
		bson.M{
			"$set": bson.M{
				"quote_pointer": bson.M{
					"$mod": bson.A{
						bson.M{"$add": bson.A{"$quote_pointer", 1}},
						totalQuotes,
					},
				},
				"updated_at": time.Now(),
			},
		},
	})
	return err
}
