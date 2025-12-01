package mongo

import (
	"context"
	"reportachievement/app/model/mongo"

	"go.mongodb.org/mongo-driver/bson"
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

// 1. Method INSERT (Dari Modul 7) - WAJIB ADA
func (r *AchievementRepository) Insert(ctx context.Context, data *mongo.Achievement) (string, error) {
	result, err := r.Coll.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	// Kembalikan ID yang baru dibuat sebagai Hex String
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// 2. Method FindByIDs (Dari Modul 8)
func (r *AchievementRepository) FindByIDs(ctx context.Context, ids []string) ([]mongo.Achievement, error) {
	var objectIDs []primitive.ObjectID

	// Convert string ID ke ObjectID Mongo
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			objectIDs = append(objectIDs, objID)
		}
	}

	// Query: WHERE _id IN (id1, id2, ...)
	filter := bson.M{"_id": bson.M{"$in": objectIDs}}

	cursor, err := r.Coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []mongo.Achievement
	if err = cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}

	return achievements, nil
}
