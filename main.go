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
		// Set HTTP headers for enhanced security
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		// w.Header().Set("Cache-Control", "max-age=0, no-cache, no-store, must-revalidate, private")
		next.ServeHTTP(w, r) // Call the next handler
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

	// CORS setup (adjust AllowedOrigins in production)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Consider specific origins in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Wrap handlers with middleware: CORS -> Security -> Logging -> Router
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

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// Get the MongoDB URI from the environment variable
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatalf("MONGODB_URI environment variable is not set")
	}

	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	// Send a ping to confirm a successful connection
	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		panic(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	BlogPostsCollection = Client.Database("eventdb").Collection("bposts")
	db.BlogPostsCollection = BlogPostsCollection
	CartCollection = Client.Database("eventdb").Collection("cart")
	db.CartCollection = CartCollection
	CatalogueCollection = Client.Database("eventdb").Collection("catalogue")
	db.CatalogueCollection = CatalogueCollection
	ChatsCollection = Client.Database("eventdb").Collection("chats")
	db.ChatsCollection = ChatsCollection
	CommentsCollection = Client.Database("eventdb").Collection("comments")
	db.CommentsCollection = CommentsCollection
	CropsCollection = Client.Database("eventdb").Collection("crops")
	db.CropsCollection = CropsCollection
	FarmsCollection = Client.Database("eventdb").Collection("farms")
	db.FarmsCollection = FarmsCollection
	FollowingsCollection = Client.Database("eventdb").Collection("followings")
	db.FollowingsCollection = FollowingsCollection
	FarmOrdersCollection = Client.Database("eventdb").Collection("forders")
	db.FarmOrdersCollection = FarmOrdersCollection
	MessagesCollection = Client.Database("eventdb").Collection("messages")
	db.MessagesCollection = MessagesCollection
	OrderCollection = Client.Database("eventdb").Collection("orders")
	db.OrderCollection = OrderCollection
	PostsCollection = Client.Database("eventdb").Collection("posts")
	db.PostsCollection = PostsCollection
	ProductCollection = Client.Database("eventdb").Collection("products")
	db.ProductCollection = ProductCollection
	ReportsCollection = Client.Database("eventdb").Collection("reports")
	db.ReportsCollection = ReportsCollection
	ReviewsCollection = Client.Database("eventdb").Collection("reviews")
	db.ReviewsCollection = ReviewsCollection
	UserCollection = Client.Database("eventdb").Collection("users")
	db.UserCollection = UserCollection
	db.Client = client

	router := httprouter.New()

	hub := newchat.NewHub()
	go hub.Run()

	rateLimiter := ratelim.NewRateLimiter()
	handler := setupRouter(rateLimiter, hub)

	router.GET("/health", Index)

	server := &http.Server{
		Addr:              ":10000",
		Handler:           handler, // Use the middleware-wrapped handler
		ReadTimeout:       7 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// Register cleanup tasks on shutdown
	server.RegisterOnShutdown(func() {
		log.Println("🛑 Cleaning up resources before shutdown...")
		// Add cleanup tasks like closing DB connection
	})

	// Start server in a goroutine to handle graceful shutdown
	go func() {
		log.Println("Server started on port 4000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port 4000: %v", err)
		}
	}()

	// Graceful shutdown on interrupt
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
