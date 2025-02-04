package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client = CreateMongoClient()

func CreateMongoClient() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading the file!")
	}
	MongoDbURI := os.Getenv("MONGODB_URI")
	if MongoDbURI == "" {
		log.Fatal("MONGODB_URI not found in environment variables")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDbURI))
	if err != nil {
		log.Fatal("Mongodb connection error: ", err)
	}

	fmt.Println("connected to Mongo Db")
	return client
}

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("go-mongodb").Collection(collectionName)

}
