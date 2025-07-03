package db

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	CartCollection       *mongo.Collection
	OrderCollection      *mongo.Collection
	CatalogueCollection  *mongo.Collection
	FarmsCollection      *mongo.Collection
	FarmOrdersCollection *mongo.Collection
	CropsCollection      *mongo.Collection
	CommentsCollection   *mongo.Collection
	UserCollection       *mongo.Collection
	ProductCollection    *mongo.Collection
	ReviewsCollection    *mongo.Collection
	FollowingsCollection *mongo.Collection
	PostsCollection      *mongo.Collection
	BlogPostsCollection  *mongo.Collection
	ChatsCollection      *mongo.Collection
	MessagesCollection   *mongo.Collection
	ReportsCollection    *mongo.Collection
	Client               *mongo.Client
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
	BlogPostsCollection = Client.Database("eventdb").Collection("bposts")
	CartCollection = Client.Database("eventdb").Collection("cart")
	CatalogueCollection = Client.Database("eventdb").Collection("catalogue")
	ChatsCollection = Client.Database("eventdb").Collection("chats")
	CommentsCollection = Client.Database("eventdb").Collection("comments")
	CropsCollection = Client.Database("eventdb").Collection("crops")
	FarmsCollection = Client.Database("eventdb").Collection("farms")
	FollowingsCollection = Client.Database("eventdb").Collection("followings")
	FarmOrdersCollection = Client.Database("eventdb").Collection("forders")
	MessagesCollection = Client.Database("eventdb").Collection("messages")
	OrderCollection = Client.Database("eventdb").Collection("orders")
	PostsCollection = Client.Database("eventdb").Collection("posts")
	ProductCollection = Client.Database("eventdb").Collection("products")
	ReportsCollection = Client.Database("eventdb").Collection("reports")
	ReviewsCollection = Client.Database("eventdb").Collection("reviews")
	UserCollection = Client.Database("eventdb").Collection("users")
}

func OptionsFindLatest(limit int64) *options.FindOptions {
	opts := options.Find()
	opts.SetSort(map[string]interface{}{"createdAt": -1})
	opts.SetLimit(limit)
	return opts
}
