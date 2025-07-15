package db

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// Your collections:
	AnalyticsCollection  *mongo.Collection
	CartCollection       *mongo.Collection
	OrderCollection      *mongo.Collection
	CatalogueCollection  *mongo.Collection
	FarmsCollection      *mongo.Collection
	FarmOrdersCollection *mongo.Collection
	CropsCollection      *mongo.Collection
	CommentsCollection   *mongo.Collection
	UserCollection       *mongo.Collection
	ProductCollection    *mongo.Collection
	UserDataCollection   *mongo.Collection
	ReviewsCollection    *mongo.Collection
	SettingsCollection   *mongo.Collection
	FollowingsCollection *mongo.Collection
	ActivitiesCollection *mongo.Collection
	ChatsCollection      *mongo.Collection
	MessagesCollection   *mongo.Collection
	ReportsCollection    *mongo.Collection
	RecipeCollection     *mongo.Collection

	Client *mongo.Client
)

// Initialize MongoDB connection
func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ClientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	Client, err = mongo.Connect(context.TODO(), ClientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// CreateIndexes(Client)
	db := Client.Database("eventdb")
	ActivitiesCollection = db.Collection("activities")
	AnalyticsCollection = db.Collection("analytics")
	CartCollection = db.Collection("cart")
	CatalogueCollection = db.Collection("catalogue")
	ChatsCollection = db.Collection("chats")
	CommentsCollection = db.Collection("comments")
	CropsCollection = db.Collection("crops")
	FarmsCollection = db.Collection("farms")
	FollowingsCollection = db.Collection("followings")
	FarmOrdersCollection = db.Collection("forders")
	MessagesCollection = db.Collection("messages")
	OrderCollection = db.Collection("orders")
	ProductCollection = db.Collection("products")
	RecipeCollection = db.Collection("recipes")
	ReportsCollection = db.Collection("reports")
	ReviewsCollection = db.Collection("reviews")
	SettingsCollection = db.Collection("settings")
	UserDataCollection = db.Collection("userdata")
	UserCollection = db.Collection("users")
}

func OptionsFindLatest(limit int64) *options.FindOptions {
	opts := options.Find()
	opts.SetSort(map[string]interface{}{"createdAt": -1})
	opts.SetLimit(limit)
	return opts
}
