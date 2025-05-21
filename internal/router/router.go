// paystack_supabase_backend/internal/router/router.go
package router

import (
	// Ensure your module name is correct in the import path
	"github.com/tedobanks/datagram_payment_processor/internal/handlers"
	"github.com/tedobanks/datagram_payment_processor/internal/middleware" // We'll create a placeholder for this

	"github.com/gin-gonic/gin"
	// For Swagger/OpenAPI documentation (optional, but good practice)
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures the Gin router with all application routes and middleware.
// It takes the initialized handlers as dependencies.
func SetupRouter(
	paymentHandler *handlers.PaymentHandler,
	// Add other handlers here if you create them, e.g.:
	// userHandler *handlers.UserHandler,
	// databyteHandler *handlers.DatabyteHandler, // If you separated databyte logic
) *gin.Engine {

	// gin.SetMode(ginMode) // Gin mode is set in main.go based on config

	router := gin.Default() // Starts with Logger and Recovery middleware by default

	// Global Middleware (example: CORS)
	// router.Use(middleware.CORSMiddleware()) // Uncomment if you implement CORS

	// Swagger documentation endpoint (optional)
	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint - useful for load balancers, uptime monitoring

	// @Summary     Health Check
	// @Description Check if the service is running
	// @Tags        system
	// @Accept      json
	// @Produce     json
	// @Success     200 {object} map[string]string
	// @Router      /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP", "message": "Datagram Payment Processor is running!"})
	})

	// API v1 group
	// All routes within this group will be prefixed with /api/v1
	apiV1 := router.Group("/api/v1")
	{
		// Payment related routes
		paymentRoutes := apiV1.Group("/payments")
		// For routes that require user authentication, you'd add an auth middleware here
		// paymentRoutes.Use(middleware.AuthMiddleware()) // Example
		{
			// Initialize a payment for datacredit purchase
			// POST /api/v1/payments/initialize

			// @Summary     Initialize Payment
			// @Description Start a new Paystack payment transaction
			// @Tags        payments
			// @Accept      json
			// @Produce     json
			// @Security    BearerAuth
			// @Param       request body handlers.InitializePaymentRequest true "Payment details"
			// @Success     200 {object} handlers.PaymentResponse
			// @Failure     400 {object} handlers.ErrorResponse
			// @Failure     401 {object} handlers.ErrorResponse
			// @Router      /payments/initialize [post]
			paymentRoutes.POST("/initialize", middleware.AuthMiddleware(), paymentHandler.InitializePayment) // Added AuthMiddleware

			// Initiate a withdrawal of datacredit
			// POST /api/v1/payments/withdraw

			// @Summary     Withdraw Funds
			// @Description Initiate a withdrawal of datacredit
			// @Tags        payments
			// @Accept      json
			// @Produce     json
			// @Security    BearerAuth
			// @Param       request body handlers.WithdrawalRequest true "Withdrawal details"
			// @Success     200 {object} handlers.WithdrawalResponse
			// @Failure     400 {object} handlers.ErrorResponse
			// @Failure     401 {object} handlers.ErrorResponse
			// @Failure     403 {object} handlers.ErrorResponse
			// @Router      /payments/withdraw [post]
			paymentRoutes.POST("/withdraw", middleware.AuthMiddleware(), paymentHandler.HandleWithdrawal) // Added AuthMiddleware
		}

		// Databyte related routes
		databyteRoutes := apiV1.Group("/databytes")
		// databyteRoutes.Use(middleware.AuthMiddleware()) // Example
		{
			// Purchase databytes using datacredit balance
			// POST /api/v1/databytes/purchase

			// @Summary     Purchase Databytes
			// @Description Purchase databytes using datacredit balance
			// @Tags        databytes
			// @Accept      json
			// @Produce     json
			// @Security    BearerAuth
			// @Param       request body handlers.DatabytePurchaseRequest true "Purchase details"
			// @Success     200 {object} handlers.DatabytePurchaseResponse
			// @Failure     400 {object} handlers.ErrorResponse
			// @Failure     401 {object} handlers.ErrorResponse
			// @Failure     403 {object} handlers.ErrorResponse
			// @Router      /databytes/purchase [post]
			databyteRoutes.POST("/purchase", middleware.AuthMiddleware(), paymentHandler.PurchaseDatabytes) // Added AuthMiddleware
		}

		// Webhook routes do NOT typically have authentication middleware,
		// as they are called by external services (Paystack).
		// Security for webhooks is handled by signature verification.
		webhookRoutes := apiV1.Group("/webhooks")
		{
			// Endpoint for Paystack to send event notifications
			// POST /api/v1/webhooks/paystack

			// @Summary     Paystack Webhook
			// @Description Endpoint for Paystack payment notifications
			// @Tags        webhooks
			// @Accept      json
			// @Produce     json
			// @Param       payload body handlers.PaystackWebhookPayload true "Webhook payload"
			// @Success     200 {object} handlers.WebhookResponse
			// @Failure     400 {object} handlers.ErrorResponse
			// @Failure     403 {object} handlers.ErrorResponse
			// @Router      /webhooks/paystack [post]
			webhookRoutes.POST("/paystack", paymentHandler.PaystackWebhook)
		}

		// Example: User profile routes (if you add user handlers)
		// profileRoutes := apiV1.Group("/profile")
		// profileRoutes.Use(middleware.AuthMiddleware())
		// {
		// 	profileRoutes.GET("", userHandler.GetMyProfile)
		// 	profileRoutes.PUT("", userHandler.UpdateMyProfile)
		//  profileRoutes.GET("/wallet", userHandler.GetMyWalletBalance)
		// }
	}

	// Fallback for unmatched routes (optional)

	// @Summary     Not Found
	// @Description Default route for unmatched paths
	// @Router      /{any} [get]
	// @Router      /{any} [post]
	// @Router      /{any} [put]
	// @Router      /{any} [delete]
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": "Not Found", "message": "The requested resource was not found on this server."})
	})

	return router
}
