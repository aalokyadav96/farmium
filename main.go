package main

import (
	"context"
	"fmt"
	"log"
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Security headers middleware
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		next.ServeHTTP(w, r)
	})
}

// MongoDB collections
var (
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
	Client               *mongo.Client
)

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "200")
}

// Set up all routes and middleware layers
func setupRouter(rateLimiter *ratelim.RateLimiter, hub *newchat.Hub) *httprouter.Router {
	router := httprouter.New()
	router.GET("/health", Index)

	routes.AddAdminRoutes(router)
	routes.AddAuthRoutes(router)
	routes.AddCartRoutes(router)
	routes.AddCommentsRoutes(router)
	routes.AddDiscordRoutes(router)
	routes.RegisterFarmRoutes(router)
	routes.AddHomeRoutes(router)
	routes.AddProfileRoutes(router)
	routes.AddRecipeRoutes(router)
	routes.AddReportRoutes(router)
	routes.AddReviewsRoutes(router)
	routes.AddSearchRoutes(router)
	routes.AddSettingsRoutes(router)
	routes.AddStaticRoutes(router)
	routes.AddSuggestionsRoutes(router)
	routes.AddUtilityRoutes(router, rateLimiter)
	routes.AddChatRoutes(router)
	routes.AddNewChatRoutes(router, hub)

	return router
}

// Simple request logging middleware
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("%s %s from %s – %v", r.Method, r.RequestURI, r.RemoteAddr, duration)
	})
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":4000"
	} else if port[0] != ':' {
		port = ":" + port
	}

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatalf("MONGODB_URI environment variable is not set")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}

	Client = client

	if err := Client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping failed: %v", err)
	}

	fmt.Println("✅ Connected to MongoDB")

	// Initialize all collections
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

	// Init rate limiter and chat hub
	rateLimiter := ratelim.NewRateLimiter()
	hub := newchat.NewHub()
	go hub.Run()

	// Setup router with all routes
	router := setupRouter(rateLimiter, hub)

	// Apply middleware
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://farmium.netlify.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(router)

	handler := loggingMiddleware(securityHeaders(corsHandler))

	server := &http.Server{
		Addr:              port,
		Handler:           handler,
		ReadTimeout:       7 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	server.RegisterOnShutdown(func() {
		log.Println("🛑 Cleaning up resources before shutdown...")
	})

	go func() {
		log.Printf("🚀 Server started on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Could not listen on port %s: %v", port, err)
		}
	}()

	// Graceful shutdown
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	<-shutdownChan

	log.Println("🛑 Shutdown signal received. Shutting down gracefully...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("❌ Server shutdown failed: %v", err)
	}

	if err := Client.Disconnect(ctxShutdown); err != nil {
		log.Printf("⚠️ MongoDB disconnect error: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
