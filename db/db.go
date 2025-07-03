package db

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	MapsCollection              *mongo.Collection
	CartCollection              *mongo.Collection
	OrderCollection             *mongo.Collection
	CatalogueCollection         *mongo.Collection
	FarmsCollection             *mongo.Collection
	FarmOrdersCollection        *mongo.Collection
	CropsCollection             *mongo.Collection
	CommentsCollection          *mongo.Collection
	RoomsCollection             *mongo.Collection
	UserCollection              *mongo.Collection
	LikesCollection             *mongo.Collection
	ProductCollection           *mongo.Collection
	ItineraryCollection         *mongo.Collection
	UserDataCollection          *mongo.Collection
	TicketsCollection           *mongo.Collection
	BehindTheScenesCollection   *mongo.Collection
	PurchasedTicketsCollection  *mongo.Collection
	ReviewsCollection           *mongo.Collection
	SettingsCollection          *mongo.Collection
	FollowingsCollection        *mongo.Collection
	PlacesCollection            *mongo.Collection
	SlotCollection              *mongo.Collection
	BookingsCollection          *mongo.Collection
	PostsCollection             *mongo.Collection
	BlogPostsCollection         *mongo.Collection
	FilesCollection             *mongo.Collection
	MerchCollection             *mongo.Collection
	MenuCollection              *mongo.Collection
	ActivitiesCollection        *mongo.Collection
	EventsCollection            *mongo.Collection
	ArtistEventsCollection      *mongo.Collection
	SongsCollection             *mongo.Collection
	MediaCollection             *mongo.Collection
	ArtistsCollection           *mongo.Collection
	CartoonsCollection          *mongo.Collection
	ChatsCollection             *mongo.Collection
	MessagesCollection          *mongo.Collection
	ReportsCollection           *mongo.Collection
	BaitoCollection             *mongo.Collection
	BaitoApplicationsCollection *mongo.Collection
	Client                      *mongo.Client
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
	ActivitiesCollection = Client.Database("eventdb").Collection("activities")
	ArtistEventsCollection = Client.Database("eventdb").Collection("artistevents")
	ArtistsCollection = Client.Database("eventdb").Collection("artists")
	BaitoCollection = Client.Database("eventdb").Collection("baito")
	BaitoApplicationsCollection = Client.Database("eventdb").Collection("baitoapply")
	BlogPostsCollection = Client.Database("eventdb").Collection("bposts")
	BookingsCollection = Client.Database("eventdb").Collection("bookings")
	BehindTheScenesCollection = Client.Database("eventdb").Collection("bts")
	CartCollection = Client.Database("eventdb").Collection("cart")
	CartoonsCollection = Client.Database("eventdb").Collection("cartoons")
	CatalogueCollection = Client.Database("eventdb").Collection("catalogue")
	ChatsCollection = Client.Database("eventdb").Collection("chats")
	CommentsCollection = Client.Database("eventdb").Collection("comments")
	CropsCollection = Client.Database("eventdb").Collection("crops")
	EventsCollection = Client.Database("eventdb").Collection("events")
	FarmsCollection = Client.Database("eventdb").Collection("farms")
	FilesCollection = Client.Database("eventdb").Collection("files")
	FollowingsCollection = Client.Database("eventdb").Collection("followings")
	FarmOrdersCollection = Client.Database("eventdb").Collection("forders")
	ItineraryCollection = Client.Database("eventdb").Collection("itinerary")
	LikesCollection = Client.Database("eventdb").Collection("likes")
	MapsCollection = Client.Database("eventdb").Collection("maps")
	MediaCollection = Client.Database("eventdb").Collection("media")
	MenuCollection = Client.Database("eventdb").Collection("menu")
	MerchCollection = Client.Database("eventdb").Collection("merch")
	MessagesCollection = Client.Database("eventdb").Collection("messages")
	OrderCollection = Client.Database("eventdb").Collection("orders")
	PlacesCollection = Client.Database("eventdb").Collection("places")
	PostsCollection = Client.Database("eventdb").Collection("posts")
	ProductCollection = Client.Database("eventdb").Collection("products")
	PurchasedTicketsCollection = Client.Database("eventdb").Collection("purticks")
	ReportsCollection = Client.Database("eventdb").Collection("reports")
	ReviewsCollection = Client.Database("eventdb").Collection("reviews")
	RoomsCollection = Client.Database("eventdb").Collection("rooms")
	SettingsCollection = Client.Database("eventdb").Collection("settings")
	SlotCollection = Client.Database("eventdb").Collection("slots")
	SongsCollection = Client.Database("eventdb").Collection("songs")
	TicketsCollection = Client.Database("eventdb").Collection("ticks")
	UserDataCollection = Client.Database("eventdb").Collection("userdata")
	UserCollection = Client.Database("eventdb").Collection("users")
}

func OptionsFindLatest(limit int64) *options.FindOptions {
	opts := options.Find()
	opts.SetSort(map[string]interface{}{"createdAt": -1})
	opts.SetLimit(limit)
	return opts
}
