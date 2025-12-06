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

// 1. Insert
func (r *AchievementRepository) Insert(ctx context.Context, data *mongo.Achievement) (string, error) {
	result, err := r.Coll.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return result.InsertedID.(primitive.ObjectID).Hex(), nil
}

// 2. FindByIDs
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

// 3. SoftDelete
func (r *AchievementRepository) SoftDelete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
	_, err = r.Coll.UpdateOne(ctx, filter, update)
	return err
}

// 4. AddAttachment
func (r *AchievementRepository) AddAttachment(ctx context.Context, id string, attachment mongo.Attachment) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$push": bson.M{"attachments": attachment}}
	_, err = r.Coll.UpdateOne(ctx, filter, update)
	return err
}

// ---  AGGREGATIONS ---

// Struct hasil agregasi Top Student
type TopStudentResult struct {
	StudentPostgresID string `bson:"_id"`
	TotalPoints       int    `bson:"totalPoints"`
	TotalAchievements int    `bson:"count"`
}

// Struct hasil agregasi Grouping (Type/Level)
type GroupStatResult struct {
	Key   string `bson:"_id"`
	Count int    `bson:"count"`
}

// A. Get Top Students (Ranking Poin)
func (r *AchievementRepository) GetTopStudents(ctx context.Context, limit int) ([]TopStudentResult, error) {
	pipeline := mongoDriver.Pipeline{
		// 1. Filter yang belum dihapus
		{{Key: "$match", Value: bson.M{"deleted_at": bson.M{"$exists": false}}}},
		// 2. Group by StudentID, Sum Points
		{{Key: "$group", Value: bson.M{
			"_id":         "$student_postgres_id",
			"totalPoints": bson.M{"$sum": "$points"},
			"count":       bson.M{"$sum": 1},
		}}},
		// 3. Sort by TotalPoints Descending
		{{Key: "$sort", Value: bson.M{"totalPoints": -1}}},
		// 4. Limit
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []TopStudentResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

// B. Get Stats by Type (Kompetisi vs Organisasi dll)
func (r *AchievementRepository) GetStatsByType(ctx context.Context) ([]GroupStatResult, error) {
	pipeline := mongoDriver.Pipeline{
		{{Key: "$match", Value: bson.M{"deleted_at": bson.M{"$exists": false}}}},
		{{Key: "$group", Value: bson.M{"_id": "$achievement_type", "count": bson.M{"$sum": 1}}}},
	}

	cursor, err := r.Coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []GroupStatResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
