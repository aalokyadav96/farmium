package main

import (
	"context"
	"fmt"
	"log"
	"naevis/db"
	"naevis/newchat"
	"naevis/ratelim"
	"naevis/routes"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Security headers middleware
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

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

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "200")
}

// Set up all routes and middleware layers
func setupRouter(rateLimiter *ratelim.RateLimiter, hub *newchat.Hub) http.Handler {
	router := httprouter.New()
	router.GET("/health", Index)

	routes.AddAuthRoutes(router)
	routes.AddCartRoutes(router)
	routes.AddChatRoutes(router)
	routes.AddCommentsRoutes(router)
	routes.RegisterFarmRoutes(router)
	routes.AddHomeRoutes(router)
	routes.AddNewChatRoutes(router, hub)
	routes.AddProfileRoutes(router)
	routes.AddReportRoutes(router)
	routes.AddReviewsRoutes(router)
	routes.AddSearchRoutes(router)
	routes.AddStaticRoutes(router)
	routes.AddSuggestionsRoutes(router)
	routes.AddUtilityRoutes(router, rateLimiter)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	return loggingMiddleware(securityHeaders(c.Handler(router)))
}

// Middleware: Simple request logging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s", r.Method, r.RequestURI, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatalf("MONGODB_URI environment variable is not set")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	Client = client // ✅ Assign to global variable

	defer func() {
		if err := Client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	if err := Client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	// Initialize all collections
	db.BlogPostsCollection = Client.Database("eventdb").Collection("bposts")
	db.CartCollection = Client.Database("eventdb").Collection("cart")
	db.CatalogueCollection = Client.Database("eventdb").Collection("catalogue")
	db.ChatsCollection = Client.Database("eventdb").Collection("chats")
	db.CommentsCollection = Client.Database("eventdb").Collection("comments")
	db.CropsCollection = Client.Database("eventdb").Collection("crops")
	db.FarmsCollection = Client.Database("eventdb").Collection("farms")
	db.FollowingsCollection = Client.Database("eventdb").Collection("followings")
	db.FarmOrdersCollection = Client.Database("eventdb").Collection("forders")
	db.MessagesCollection = Client.Database("eventdb").Collection("messages")
	db.OrderCollection = Client.Database("eventdb").Collection("orders")
	db.PostsCollection = Client.Database("eventdb").Collection("posts")
	db.ProductCollection = Client.Database("eventdb").Collection("products")
	db.ReportsCollection = Client.Database("eventdb").Collection("reports")
	db.ReviewsCollection = Client.Database("eventdb").Collection("reviews")
	db.UserCollection = Client.Database("eventdb").Collection("users")
	db.Client = Client

	// Assign global collections for this package (if needed)
	BlogPostsCollection = db.BlogPostsCollection
	CartCollection = db.CartCollection
	CatalogueCollection = db.CatalogueCollection
	ChatsCollection = db.ChatsCollection
	CommentsCollection = db.CommentsCollection
	CropsCollection = db.CropsCollection
	FarmsCollection = db.FarmsCollection
	FollowingsCollection = db.FollowingsCollection
	FarmOrdersCollection = db.FarmOrdersCollection
	MessagesCollection = db.MessagesCollection
	OrderCollection = db.OrderCollection
	PostsCollection = db.PostsCollection
	ProductCollection = db.ProductCollection
	ReportsCollection = db.ReportsCollection
	ReviewsCollection = db.ReviewsCollection
	UserCollection = db.UserCollection

	hub := newchat.NewHub()
	go hub.Run()

	rateLimiter := ratelim.NewRateLimiter()
	handler := setupRouter(rateLimiter, hub)

	server := &http.Server{
		Addr:              ":10000",
		Handler:           handler,
		ReadTimeout:       7 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	server.RegisterOnShutdown(func() {
		log.Println("🛑 Cleaning up resources before shutdown...")
		// You may add cleanup tasks here if needed
	})

	go func() {
		log.Println("Server started on port 10000") // ✅ Fixed log
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port 10000: %v", err)
		}
	}()

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	<-shutdownChan

	log.Println("🛑 Shutdown signal received. Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
