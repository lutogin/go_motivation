package mongo

import (
	"context"

	"github.com/aluto/go-motivation/internal/entity"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type QuoteRepo struct {
	col *mongo.Collection
}

func NewQuoteRepo(db *mongo.Database) *QuoteRepo {
	return &QuoteRepo{col: db.Collection("quotes")}
}

func (r *QuoteRepo) Insert(ctx context.Context, q *entity.Quote) error {
	_, err := r.col.InsertOne(ctx, q)
	return err
}

func (r *QuoteRepo) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}

func (r *QuoteRepo) GetByIndex(ctx context.Context, index int) (*entity.Quote, error) {
	opts := options.FindOne().SetSkip(int64(index)).SetSort(bson.D{{Key: "_id", Value: 1}})
	var q entity.Quote
	if err := r.col.FindOne(ctx, bson.M{}, opts).Decode(&q); err != nil {
		return nil, err
	}
	return &q, nil
}
