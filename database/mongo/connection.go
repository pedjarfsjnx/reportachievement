package mongo

import (
	"context"
	"log"
	"reportachievement/config" // Import Config
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

func Connect(cfg *config.Config) MongoInstance {
	// Gunakan URI dari Config
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Gagal koneksi ke MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ Gagal Ping MongoDB:", err)
	}

	log.Println("✅ Terkoneksi ke MongoDB!")

	// Gunakan DB Name dari Config
	db := client.Database(cfg.MongoDBName)

	return MongoInstance{
		Client: client,
		Db:     db,
	}
}
