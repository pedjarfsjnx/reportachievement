package mongo // Pastikan nama package ini benar

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

func Connect() MongoInstance {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Pastikan variable MONGO_URI ada di .env
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Gagal membuat client MongoDB: ", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ Gagal ping ke MongoDB: ", err)
	}

	log.Println("✅ Berhasil koneksi ke MongoDB!")

	return MongoInstance{
		Client: client,
		Db:     client.Database(os.Getenv("MONGO_DB_NAME")),
	}
}
