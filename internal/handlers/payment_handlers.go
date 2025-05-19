// package handlers

// import (
// 	"encoding/json"
// 	// "fmt" // Added for potential error messages
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"strings" // For error message checking (use proper error types in production)

// 	// Internal application packages (replace 'your_module_name')
// 	"github.com/tedobanks/datagram_payment_processor/internal/services"
// 	"github.com/tedobanks/datagram_payment_processor/internal/models"
// 	"github.com/tedobanks/datagram_payment_processor/internal/utils"

// 	"github.com/gin-gonic/gin"
// )

// // PaymentHandler holds dependencies for payment and currency-related handlers
// type PaymentHandler struct {
// 	PaystackService *services.PaystackService
// 	SupabaseService *services.SupabaseService
// 	// Add other services if needed, e.g., a dedicated DatabyteService if logic grows
// }

// // NewPaymentHandler creates a new PaymentHandler
// func NewPaymentHandler(ps *services.PaystackService, ss *services.SupabaseService) *PaymentHandler {
// 	return &PaymentHandler{
// 		PaystackService: ps,
// 		SupabaseService: ss,
// 	}
// }

// // InitializePayment godoc
// // @Summary Initialize a payment transaction
// // @Description Initializes a payment with Paystack for datacredit purchase.
// // @Tags Payments
// // @Accept  json
// // @Produce  json
// // @Param   paymentRequest body models.PaystackInitializeRequest true "Payment Request"
// // @Success 200 {object} map[string]interface{} "message, authorization_url, access_code, reference"
// // @Failure 400 {object} utils.ErrorResponse "Bad Request: Invalid input"
// // @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// // @Router /payments/initialize [post]
// func (h *PaymentHandler) InitializePayment(c *gin.Context) {
// 	var req models.PaystackInitializeRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
// 		return
// 	}

// 	// CRITICAL: Get UserID from authenticated session/JWT.
// 	// This is a placeholder. In a real application, an authentication middleware
// 	// should verify the user and make their ID available in the Gin context.
// 	userIDFromAuth, exists := c.Get("userID") // Example: if auth middleware sets "userID"
// 	if !exists {
// 		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
// 		return
// 	}
// 	userID := userIDFromAuth.(string) // Type assert, ensure your auth middleware stores it as string

// 	// You might want to add a callback URL that Paystack redirects to after payment attempt
// 	// This URL could be configured or dynamically generated.
// 	// For example: "https://yourapp.com/payment/callback"
// 	callbackURL := "YOUR_APPLICATION_PAYMENT_CALLBACK_URL" // Replace with actual or configured URL

// 	resp, err := h.PaystackService.InitializePayment(req, userID, callbackURL)
// 	if err != nil {
// 		log.Printf("Error initializing payment for UserID %s: %v", userID, err)
// 		// Check if the error message indicates a Paystack specific issue known from SDK
// 		if strings.Contains(err.Error(), "Paystack initialization failed") {
// 			utils.RespondWithError(c, http.StatusServiceUnavailable, err.Error())
// 		} else {
// 			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to initialize payment")
// 		}
// 		return
// 	}

// 	utils.RespondWithJSON(c, http.StatusOK, gin.H{
// 		"message":           "Payment initialization successful",
// 		"authorization_url": resp.Data.AuthorizationURL,
// 		"access_code":       resp.Data.AccessCode,
// 		"reference":         resp.Data.Reference,
// 	})
// }

// // PaystackWebhook godoc
// // @Summary Handles incoming webhooks from Paystack
// // @Description Verifies and processes Paystack webhook events, e.g., crediting users on successful payment.
// // @Tags Webhooks
// // @Accept  json
// // @Produce  json
// // @Param X-Paystack-Signature header string true "Paystack Signature from Paystack"
// // @Success 200 {object} map[string]string "status: 'Webhook processed'"
// // @Failure 400 {object} utils.ErrorResponse "Bad Request or Signature Verification Failed"
// // @Failure 401 {object} utils.ErrorResponse "Unauthorized (Signature Verification Failed)"
// // @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// // @Router /webhooks/paystack [post]
// func (h *PaymentHandler) PaystackWebhook(c *gin.Context) {
// 	body, err := ioutil.ReadAll(c.Request.Body)
// 	if err != nil {
// 		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to read request body")
// 		return
// 	}
// 	defer c.Request.Body.Close() // Important to close the body

// 	signature := c.GetHeader("X-Paystack-Signature")
// 	if signature == "" {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Missing X-Paystack-Signature header")
// 		return
// 	}

// 	if !h.PaystackService.VerifyWebhookSignature(body, signature) {
// 		log.Println("Paystack webhook signature verification failed.")
// 		// Use 401 Unauthorized if signature verification fails, as it implies the request isn't trusted.
// 		utils.RespondWithError(c, http.StatusUnauthorized, "Webhook signature verification failed")
// 		return
// 	}

// 	var payload models.PaystackWebhookPayload
// 	if err := json.Unmarshal(body, &payload); err != nil {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Failed to parse webhook payload: "+err.Error())
// 		return
// 	}

// 	log.Printf("Received verified Paystack webhook. Event: %s", payload.Event)

// 	switch payload.Event {
// 	case "charge.success":
// 		var transactionData models.PaystackTransactionData
// 		dataBytes, _ := json.Marshal(payload.Data) // Convert interface{} to bytes
// 		if err := json.Unmarshal(dataBytes, &transactionData); err != nil {
// 			log.Printf("Error unmarshalling charge.success data: %v", err)
// 			utils.RespondWithError(c, http.StatusInternalServerError, "Error processing charge.success event data")
// 			return
// 		}

// 		err := h.PaystackService.ProcessSuccessfulPayment(transactionData, h.SupabaseService)
// 		if err != nil {
// 			log.Printf("Error processing successful payment for Paystack reference %s: %v", transactionData.Reference, err)
// 			utils.RespondWithError(c, http.StatusInternalServerError, "Error processing successful payment")
// 			return
// 		}
// 		log.Printf("Successfully processed charge.success for Paystack reference: %s", transactionData.Reference)

// 	case "transfer.success":
// 		log.Printf("Received transfer.success webhook: %+v", payload.Data)
// 		// TODO: Implement logic for successful transfers
// 		// e.g., final logging, notifying user that withdrawal is complete.
// 		// This is where you'd be certain the money has reached the user.

// 	case "transfer.failed":
// 		log.Printf("Received transfer.failed webhook: %+v", payload.Data)
// 		// TODO: Implement logic for failed transfers
// 		// e.g., Re-credit user if debited prematurely, notify user, investigate failure.

// 	case "transfer.reversed":
// 		log.Printf("Received transfer.reversed webhook: %+v", payload.Data)
// 		// TODO: Implement logic for reversed transfers
// 		// This might happen after a transfer was initially successful.

// 	default:
// 		log.Printf("Unhandled Paystack webhook event: %s", payload.Event)
// 	}

// 	utils.RespondWithJSON(c, http.StatusOK, gin.H{"status": "Webhook processed"})
// }

// // HandleWithdrawal godoc
// // @Summary Initiate a datacredit withdrawal for a user
// // @Description Debits user's datacredit balance and attempts to initiate a transfer via Paystack.
// // @Tags Payments
// // @Accept  json
// // @Produce  json
// // @Param   withdrawalRequest body models.WithdrawalRequest true "Withdrawal Request"
// // @Success 200 {object} map[string]string "message: 'Withdrawal initiated', transfer_code: '...'"
// // @Failure 400 {object} utils.ErrorResponse "Bad Request: Invalid input or insufficient datacredits"
// // @Failure 401 {object} utils.ErrorResponse "User not authenticated"
// // @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// // @Router /payments/withdraw [post]
// func (h *PaymentHandler) HandleWithdrawal(c *gin.Context) {
// 	var req models.WithdrawalRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
// 		return
// 	}

// 	// CRITICAL: Get UserID from authenticated session/JWT.
// 	// Ensure the UserID in the request matches the authenticated user or is authorized.
// 	// For simplicity, we're using req.UserID, but in a real app, this needs robust auth.
// 	userIDFromAuth, exists := c.Get("userID")
// 	if !exists || userIDFromAuth.(string) != req.UserID { // Assuming userID in JWT is string
// 		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated or mismatched UserID")
// 		return
// 	}
// 	// UserID is now validated against authenticated user.

// 	transferCode, err := h.PaystackService.InitiateWithdrawal(req, h.SupabaseService)
// 	if err != nil {
// 		log.Printf("Error initiating withdrawal for UserID %s: %v", req.UserID, err)
// 		// Using strings.Contains is fragile. Ideally, services.PaystackService.InitiateWithdrawal
// 		// would return specific error types (e.g., ErrInsufficientBalance).
// 		if strings.Contains(strings.ToLower(err.Error()), "insufficient datacredit balance") {
// 			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
// 		} else if strings.Contains(strings.ToLower(err.Error()), "paystack transfer initiation failed") {
// 			utils.RespondWithError(c, http.StatusServiceUnavailable, err.Error()) // Paystack specific issue
// 		} else {
// 			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to initiate withdrawal: "+err.Error())
// 		}
// 		return
// 	}

// 	utils.RespondWithJSON(c, http.StatusOK, gin.H{
// 		"message":       "Withdrawal initiation request processed", // Message reflects that it's an async process
// 		"transfer_code": transferCode,
// 	})
// }

// // PurchaseDatabytes godoc
// // @Summary Purchase Databytes using Datacredit balance
// // @Description Converts a user's Datacredit balance into Databytes based on the configured exchange rate.
// // @Tags Databytes
// // @Accept  json
// // @Produce  json
// // @Param   purchaseRequest body models.DatabytePurchaseRequest true "Databyte Purchase Request"
// // @Success 200 {object} models.Wallet "Updated wallet information reflecting the purchase"
// // @Failure 400 {object} utils.ErrorResponse "Bad Request: Invalid input or insufficient datacredits"
// // @Failure 401 {object} utils.ErrorResponse "User not authenticated"
// // @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// // @Router /databytes/purchase [post]
// func (h *PaymentHandler) PurchaseDatabytes(c *gin.Context) {
// 	var req models.DatabytePurchaseRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
// 		return
// 	}

// 	// CRITICAL: Get UserID from authenticated session/JWT.
// 	// Ensure the UserID in the request matches the authenticated user.
// 	userIDFromAuth, exists := c.Get("userID")
// 	if !exists || userIDFromAuth.(string) != req.UserID { // Assuming userID in JWT is string
// 		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated or mismatched UserID")
// 		return
// 	}
// 	// UserID is now validated against authenticated user.

// 	if req.DatabyteAmount <= 0 {
// 		utils.RespondWithError(c, http.StatusBadRequest, "Databyte amount must be positive")
// 		return
// 	}

// 	wallet, err := h.SupabaseService.PurchaseDatabytesWithDatacredit(req.UserID, req.DatabyteAmount)
// 	if err != nil {
// 		// Again, using strings.Contains is not ideal.
// 		// Define custom error types in your service layer (e.g., services.ErrInsufficientDatacredit).
// 		if strings.Contains(strings.ToLower(err.Error()), "insufficient datacredit balance") {
// 			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
// 		} else if strings.Contains(strings.ToLower(err.Error()), "invalid configuration for databytes_per_datacredit_kobo") {
// 			log.Printf("Configuration error during databyte purchase for UserID %s: %v", req.UserID, err)
// 			utils.RespondWithError(c, http.StatusInternalServerError, "Configuration error, please contact support.")
// 		} else {
// 			log.Printf("Error purchasing databytes for UserID %s: %v", req.UserID, err)
// 			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to purchase databytes")
// 		}
// 		return
// 	}

// 	utils.RespondWithJSON(c, http.StatusOK, wallet)
// }

package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/tedobanks/datagram_payment_processor/internal/services"
	"github.com/tedobanks/datagram_payment_processor/internal/models"
	"github.com/tedobanks/datagram_payment_processor/internal/utils"

	"github.com/gin-gonic/gin"
)

// PaymentHandler holds dependencies for payment and currency-related handlers
type PaymentHandler struct {
	PaystackService *services.PaystackService
	SupabaseService *services.SupabaseService
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(ps *services.PaystackService, ss *services.SupabaseService) *PaymentHandler {
	return &PaymentHandler{
		PaystackService: ps,
		SupabaseService: ss,
	}
}

// InitializePayment godoc
// @Summary Initialize a payment transaction
// @Description Initializes a payment with Paystack for datacredit purchase.
// @Tags Payments
// @Accept  json
// @Produce  json
// @Param   paymentRequest body models.PaystackInitializeRequest true "Payment Request"
// @Success 200 {object} map[string]interface{} "message, authorization_url, access_code, reference"
// @Failure 400 {object} utils.ErrorResponse "Bad Request: Invalid input"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /payments/initialize [post]
func (h *PaymentHandler) InitializePayment(c *gin.Context) {
	var req models.PaystackInitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// CRITICAL: Get UserID from authenticated session/JWT.
	// This is a placeholder. In a real application, an authentication middleware
	// should verify the user and make their ID available in the Gin context.
	userIDFromAuth, exists := c.Get("userID") // Example: if auth middleware sets "userID"
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	userID := userIDFromAuth.(string) // Type assert, ensure your auth middleware stores it as string

	// You might want to add a callback URL that Paystack redirects to after payment attempt
	// This URL could be configured or dynamically generated.
	// For example: "https://yourapp.com/payment/callback"
	callbackURL := "YOUR_APPLICATION_PAYMENT_CALLBACK_URL" // Replace with actual or configured URL

	resp, err := h.PaystackService.InitializePayment(req, userID, callbackURL)
	if err != nil {
		log.Printf("Error initializing payment for UserID %s: %v", userID, err)
		// Check if the error message indicates a Paystack specific issue known from SDK
		if strings.Contains(err.Error(), "Paystack initialization failed") {
			utils.RespondWithError(c, http.StatusServiceUnavailable, err.Error())
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to initialize payment")
		}
		return
	}

	// The Paystack Go SDK typically returns a response with different structure
	// Check the actual response structure from your PaystackService.InitializePayment method
	// It might return a custom response type or the response might have different field names
	
	// Assuming the service returns a custom response type with the needed fields
	// You'll need to adjust this based on what your PaystackService.InitializePayment actually returns
	utils.RespondWithJSON(c, http.StatusOK, gin.H{
		"message": "Payment initialization successful",
		"data":    resp, // Return the entire response from the service
	})
}

// PaystackWebhook godoc
// @Summary Handles incoming webhooks from Paystack
// @Description Verifies and processes Paystack webhook events, e.g., crediting users on successful payment.
// @Tags Webhooks
// @Accept  json
// @Produce  json
// @Param X-Paystack-Signature header string true "Paystack Signature from Paystack"
// @Success 200 {object} map[string]string "status: 'Webhook processed'"
// @Failure 400 {object} utils.ErrorResponse "Bad Request or Signature Verification Failed"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized (Signature Verification Failed)"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /webhooks/paystack [post]
func (h *PaymentHandler) PaystackWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to read request body")
		return
	}
	defer c.Request.Body.Close() // Important to close the body

	signature := c.GetHeader("X-Paystack-Signature")
	if signature == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "Missing X-Paystack-Signature header")
		return
	}

	if !h.PaystackService.VerifyWebhookSignature(body, signature) {
		log.Println("Paystack webhook signature verification failed.")
		// Use 401 Unauthorized if signature verification fails, as it implies the request isn't trusted.
		utils.RespondWithError(c, http.StatusUnauthorized, "Webhook signature verification failed")
		return
	}

	var payload models.PaystackWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to parse webhook payload: "+err.Error())
		return
	}

	log.Printf("Received verified Paystack webhook. Event: %s", payload.Event)

	switch payload.Event {
	case "charge.success":
		var transactionData models.PaystackTransactionData
		dataBytes, _ := json.Marshal(payload.Data) // Convert interface{} to bytes
		if err := json.Unmarshal(dataBytes, &transactionData); err != nil {
			log.Printf("Error unmarshalling charge.success data: %v", err)
			utils.RespondWithError(c, http.StatusInternalServerError, "Error processing charge.success event data")
			return
		}

		err := h.PaystackService.ProcessSuccessfulPayment(transactionData, h.SupabaseService)
		if err != nil {
			log.Printf("Error processing successful payment for Paystack reference %s: %v", transactionData.Reference, err)
			utils.RespondWithError(c, http.StatusInternalServerError, "Error processing successful payment")
			return
		}
		log.Printf("Successfully processed charge.success for Paystack reference: %s", transactionData.Reference)

	case "transfer.success":
		log.Printf("Received transfer.success webhook: %+v", payload.Data)
		// TODO: Implement logic for successful transfers
		// e.g., final logging, notifying user that withdrawal is complete.
		// This is where you'd be certain the money has reached the user.

	case "transfer.failed":
		log.Printf("Received transfer.failed webhook: %+v", payload.Data)
		// TODO: Implement logic for failed transfers
		// e.g., Re-credit user if debited prematurely, notify user, investigate failure.

	case "transfer.reversed":
		log.Printf("Received transfer.reversed webhook: %+v", payload.Data)
		// TODO: Implement logic for reversed transfers
		// This might happen after a transfer was initially successful.

	default:
		log.Printf("Unhandled Paystack webhook event: %s", payload.Event)
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{"status": "Webhook processed"})
}

// HandleWithdrawal godoc
// @Summary Initiate a datacredit withdrawal for a user
// @Description Debits user's datacredit balance and attempts to initiate a transfer via Paystack.
// @Tags Payments
// @Accept  json
// @Produce  json
// @Param   withdrawalRequest body models.WithdrawalRequest true "Withdrawal Request"
// @Success 200 {object} map[string]string "message: 'Withdrawal initiated', transfer_code: '...'"
// @Failure 400 {object} utils.ErrorResponse "Bad Request: Invalid input or insufficient datacredits"
// @Failure 401 {object} utils.ErrorResponse "User not authenticated"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /payments/withdraw [post]
func (h *PaymentHandler) HandleWithdrawal(c *gin.Context) {
	var req models.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// CRITICAL: Get UserID from authenticated session/JWT.
	// Ensure the UserID in the request matches the authenticated user or is authorized.
	// For simplicity, we're using req.UserID, but in a real app, this needs robust auth.
	userIDFromAuth, exists := c.Get("userID")
	if !exists || userIDFromAuth.(string) != req.UserID { // Assuming userID in JWT is string
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated or mismatched UserID")
		return
	}
	// UserID is now validated against authenticated user.

	transferCode, err := h.PaystackService.InitiateWithdrawal(req, h.SupabaseService)
	if err != nil {
		log.Printf("Error initiating withdrawal for UserID %s: %v", req.UserID, err)
		// Using strings.Contains is fragile. Ideally, services.PaystackService.InitiateWithdrawal
		// would return specific error types (e.g., ErrInsufficientBalance).
		if strings.Contains(strings.ToLower(err.Error()), "insufficient datacredit balance") {
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		} else if strings.Contains(strings.ToLower(err.Error()), "paystack transfer initiation failed") {
			utils.RespondWithError(c, http.StatusServiceUnavailable, err.Error()) // Paystack specific issue
		} else {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to initiate withdrawal: "+err.Error())
		}
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, gin.H{
		"message":       "Withdrawal initiation request processed", // Message reflects that it's an async process
		"transfer_code": transferCode,
	})
}

// PurchaseDatabytes godoc
// @Summary Purchase Databytes using Datacredit balance
// @Description Converts a user's Datacredit balance into Databytes based on the configured exchange rate.
// @Tags Databytes
// @Accept  json
// @Produce  json
// @Param   purchaseRequest body models.DatabytePurchaseRequest true "Databyte Purchase Request"
// @Success 200 {object} models.Wallet "Updated wallet information reflecting the purchase"
// @Failure 400 {object} utils.ErrorResponse "Bad Request: Invalid input or insufficient datacredits"
// @Failure 401 {object} utils.ErrorResponse "User not authenticated"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /databytes/purchase [post]
func (h *PaymentHandler) PurchaseDatabytes(c *gin.Context) {
	var req models.DatabytePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}

	// CRITICAL: Get UserID from authenticated session/JWT.
	// Ensure the UserID in the request matches the authenticated user.
	userIDFromAuth, exists := c.Get("userID")
	if !exists || userIDFromAuth.(string) != req.UserID { // Assuming userID in JWT is string
		utils.RespondWithError(c, http.StatusUnauthorized, "User not authenticated or mismatched UserID")
		return
	}
	// UserID is now validated against authenticated user.

	if req.DatabyteAmount <= 0 {
		utils.RespondWithError(c, http.StatusBadRequest, "Databyte amount must be positive")
		return
	}

	wallet, err := h.SupabaseService.PurchaseDatabytesWithDatacredit(req.UserID, req.DatabyteAmount)
	if err != nil {
		// Again, using strings.Contains is not ideal.
		// Define custom error types in your service layer (e.g., services.ErrInsufficientDatacredit).
		if strings.Contains(strings.ToLower(err.Error()), "insufficient datacredit balance") {
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
		} else if strings.Contains(strings.ToLower(err.Error()), "invalid configuration for databytes_per_datacredit_kobo") {
			log.Printf("Configuration error during databyte purchase for UserID %s: %v", req.UserID, err)
			utils.RespondWithError(c, http.StatusInternalServerError, "Configuration error, please contact support.")
		} else {
			log.Printf("Error purchasing databytes for UserID %s: %v", req.UserID, err)
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to purchase databytes")
		}
		return
	}

	utils.RespondWithJSON(c, http.StatusOK, wallet)
}