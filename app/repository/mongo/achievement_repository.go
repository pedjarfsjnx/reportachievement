package mongo

import (
	"context"
	"reportachievement/app/model/mongo"

	"go.mongodb.org/mongo-driver/bson/primitive"
	mongoDriver "go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository struct {
	Coll *mongoDriver.Collection
}

func NewAchievementRepository(db *mongoDriver.Database) *AchievementRepository {
	return &AchievementRepository{
		Coll: db.Collection("achievements"),
	}
}

func (r *AchievementRepository) Insert(ctx context.Context, data *mongo.Achievement) (string, error) {
	// Insert ke MongoDB
	result, err := r.Coll.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	// Kembalikan ID yang baru dibuat sebagai Hex String
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}
