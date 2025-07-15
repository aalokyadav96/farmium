package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"naevis/newchat"
	"naevis/ratelim"
	"naevis/routes"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

// securityHeaders applies a set of recommended HTTP security headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// XSS, content sniffing, framing
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		// HSTS (must be on HTTPS)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		// Referrer and permissions
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		// Prevent caching
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs each request method, path, remote address, and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		log.Printf("%s %s from %s – %v", r.Method, r.RequestURI, r.RemoteAddr, duration)
	})
}

// Index is a simple health check handler.
func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	fmt.Fprint(w, "200")
}

// setupRouter builds the router with all routes except chat.
// The chat routes will be added separately in main to avoid passing hub around globally.
func setupRouter(rateLimiter *ratelim.RateLimiter) *httprouter.Router {
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

	return router
}

func main() {
	// load .env if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment")
	}

	// read port
	port := os.Getenv("PORT")
	if port == "" {
		port = ":10000"
	} else if port[0] != ':' {
		port = ":" + port
	}

	// initialize rate limiter
	rateLimiter := ratelim.NewRateLimiter()

	// initialize chat hub
	hub := newchat.NewHub()
	go hub.Run()

	// build router and add chat routes with hub
	router := setupRouter(rateLimiter)
	routes.AddChatRoutes(router)         // existing chat routes without hub
	routes.AddNewChatRoutes(router, hub) // newchat routes that need hub

	// apply middleware: CORS → security headers → logging → router
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // lock down in production
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(router)

	handler := loggingMiddleware(securityHeaders(corsHandler))

	// create HTTP server
	server := &http.Server{
		Addr:              port,
		Handler:           handler,
		ReadTimeout:       7 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	// on shutdown: stop chat hub, cleanup
	server.RegisterOnShutdown(func() {
		log.Println("🛑 Shutting down chat hub...")
		hub.Stop() // implement Stop() in newchat.Hub to close all connections
		// close DB connections, flush logs, etc.
	})

	// start server
	go func() {
		log.Printf("🚀 Server listening on %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ ListenAndServe error: %v", err)
		}
	}()

	// wait for interrupt or SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	// initiate graceful shutdown
	log.Println("🛑 Shutdown signal received; shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Graceful shutdown failed: %v", err)
	}

	log.Println("✅ Server stopped cleanly")
}
