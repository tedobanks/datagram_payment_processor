// paystack_supabase_backend/main.go
package main

import (
	"fmt"
	"log"
	// "os" // For signal handling

	// Ensure your module name is correct in these import paths
	"github.com/tedobanks/datagram_payment_processor/internal/config"
	"github.com/tedobanks/datagram_payment_processor/internal/handlers"
	"github.com/tedobanks/datagram_payment_processor/internal/router"
	"github.com/tedobanks/datagram_payment_processor/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/tedobanks/datagram_payment_processor/docs" 
	// "net/http" // For http.Server if doing graceful shutdown
	// "os/signal" // For signal handling
	// "syscall"   // For signal handling
	// "time"      // For graceful shutdown timeout
)

// --- General API Information for Swagger ---
// @title           Datagram Payment Processor API
// @version         1.0
// @description     Backend API for handling Paystack payments, user credits, and databyte conversions.
// @termsOfService  http://example.com/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /
// @schemes   http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// 1. Load Application Configuration
	// This will load from .env file (if present locally) or environment variables.
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("FATAL: Failed to load configuration: %v", err)
	}

	// 2. Initialize Services
	// These services encapsulate business logic and interactions with external systems (Supabase, Paystack).

	// Initialize Supabase Service
	supabaseService, err := services.NewSupabaseService(cfg)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Supabase service: %v", err)
	}

	// Initialize Paystack Service
	paystackService, err := services.NewPaystackService(cfg)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Paystack service: %v", err)
	}

	// Initialize other services here if you add more (e.g., a dedicated UserService).

	// 3. Initialize HTTP Handlers
	// Handlers take services as dependencies and process HTTP requests.
	paymentHandler := handlers.NewPaymentHandler(paystackService, supabaseService)
	// userHandler := handlers.NewUserHandler(supabaseService) // Example if you create UserHandler

	// 4. Setup Gin Router
	// The router defines API endpoints and maps them to handlers.
	// Set Gin mode (debug, release, test) based on configuration.
	gin.SetMode(cfg.GinMode)
	
	// Pass all initialized handlers to the router setup function.
	appRouter := router.SetupRouter(paymentHandler /*, userHandler */)

	// 5. Start HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("INFO: Server starting on %s (Gin Mode: %s)", serverAddr, cfg.GinMode)

	// To implement graceful shutdown (recommended for production):
	// srv := &http.Server{
	// 	Addr:    serverAddr,
	// 	Handler: appRouter,
	// }
	// go func() {
	// 	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("FATAL: ListenAndServe error: %s\n", err)
	// 	}
	// }()
	// // Wait for interrupt signal to gracefully shut down the server
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// <-quit
	// log.Println("INFO: Shutting down server...")
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 5-second timeout for shutdown
	// defer cancel()
	// if err := srv.Shutdown(ctx); err != nil {
	// 	log.Fatalf("FATAL: Server forced to shutdown: %v", err)
	// }
	// log.Println("INFO: Server exiting.")

	// Simpler startup (without graceful shutdown for now):
	if err := appRouter.Run(serverAddr); err != nil {
		log.Fatalf("FATAL: Failed to start Gin server: %v", err)
	}
}
