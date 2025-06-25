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
	BaitoApplicationsCollection *mongo.Collection
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
	ForumsCollection            *mongo.Collection
	ChatsCollection             *mongo.Collection
	MessagesCollection          *mongo.Collection
	ReportsCollection           *mongo.Collection
	BaitoCollection             *mongo.Collection
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
	MapsCollection = Client.Database("eventdb").Collection("maps")
	OrderCollection = Client.Database("eventdb").Collection("orders")
	CartCollection = Client.Database("eventdb").Collection("cart")
	BaitoApplicationsCollection = Client.Database("eventdb").Collection("baitoapply")
	CatalogueCollection = Client.Database("eventdb").Collection("catalogue")
	FarmsCollection = Client.Database("eventdb").Collection("farms")
	FarmOrdersCollection = Client.Database("eventdb").Collection("forders")
	CropsCollection = Client.Database("eventdb").Collection("crops")
	CommentsCollection = Client.Database("eventdb").Collection("comments")
	RoomsCollection = Client.Database("eventdb").Collection("rooms")
	SettingsCollection = Client.Database("eventdb").Collection("settings")
	ReviewsCollection = Client.Database("eventdb").Collection("reviews")
	FollowingsCollection = Client.Database("eventdb").Collection("followings")
	LikesCollection = Client.Database("eventdb").Collection("likes")
	ProductCollection = Client.Database("eventdb").Collection("products")
	ItineraryCollection = Client.Database("eventdb").Collection("itinerary")
	UserCollection = Client.Database("eventdb").Collection("users")
	UserDataCollection = Client.Database("eventdb").Collection("userdata")
	TicketsCollection = Client.Database("eventdb").Collection("ticks")
	PurchasedTicketsCollection = Client.Database("eventdb").Collection("purticks")
	BehindTheScenesCollection = Client.Database("eventdb").Collection("bts")
	PlacesCollection = Client.Database("eventdb").Collection("places")
	BookingsCollection = Client.Database("eventdb").Collection("bookings")
	SlotCollection = Client.Database("eventdb").Collection("slots")
	PostsCollection = Client.Database("eventdb").Collection("posts")
	FilesCollection = Client.Database("eventdb").Collection("files")
	MerchCollection = Client.Database("eventdb").Collection("merch")
	MenuCollection = Client.Database("eventdb").Collection("menu")
	ActivitiesCollection = Client.Database("eventdb").Collection("activities")
	EventsCollection = Client.Database("eventdb").Collection("events")
	ArtistEventsCollection = Client.Database("eventdb").Collection("artistevents")
	SongsCollection = Client.Database("eventdb").Collection("songs")
	MediaCollection = Client.Database("eventdb").Collection("media")
	ArtistsCollection = Client.Database("eventdb").Collection("artists")
	CartoonsCollection = Client.Database("eventdb").Collection("cartoons")
	ChatsCollection = Client.Database("eventdb").Collection("chats")
	ForumsCollection = Client.Database("eventdb").Collection("forums")
	MessagesCollection = Client.Database("eventdb").Collection("messages")
	ReportsCollection = Client.Database("eventdb").Collection("reports")
	BaitoCollection = Client.Database("eventdb").Collection("baito")
}

func OptionsFindLatest(limit int64) *options.FindOptions {
	opts := options.Find()
	opts.SetSort(map[string]interface{}{"createdAt": -1})
	opts.SetLimit(limit)
	return opts
}
