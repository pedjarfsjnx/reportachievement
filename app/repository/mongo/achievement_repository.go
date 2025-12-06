package mongo

import (
	"context"
	"reportachievement/app/model/mongo"
	"time"

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

// 1. Method INSERT
func (r *AchievementRepository) Insert(ctx context.Context, data *mongo.Achievement) (string, error) {
	result, err := r.Coll.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// 2. Method FindByIDs
func (r *AchievementRepository) FindByIDs(ctx context.Context, ids []string) ([]mongo.Achievement, error) {
	var objectIDs []primitive.ObjectID

	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			objectIDs = append(objectIDs, objID)
		}
	}

	filter := bson.M{
		"_id":        bson.M{"$in": objectIDs},
		"deleted_at": bson.M{"$exists": false},
	}

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

// 3. Method SoftDelete
func (r *AchievementRepository) SoftDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"deleted_at": time.Now(),
		},
	}

	_, err = r.Coll.UpdateOne(ctx, filter, update)
	return err
}

// Push Attachment ---
func (r *AchievementRepository) AddAttachment(ctx context.Context, id string, attachment mongo.Attachment) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objID}
	// Gunakan $push untuk menambah item ke array attachments
	update := bson.M{
		"$push": bson.M{
			"attachments": attachment,
		},
	}

	_, err = r.Coll.UpdateOne(ctx, filter, update)
	return err
}
