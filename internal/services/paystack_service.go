package services

import (
	// Standard library imports
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"log"

	// Third-party imports
	"github.com/rpip/paystack-go"

	"github.com/tedobanks/datagram_payment_processor/internal/config"
	"github.com/tedobanks/datagram_payment_processor/internal/models"
)

// PaystackService provides methods to interact with Paystack
type PaystackService struct {
	Client *paystack.Client
	Cfg    *config.Config // Store config for access to secret key for webhooks etc.
}

// NewPaystackService creates a new PaystackService
func NewPaystackService(cfg *config.Config) (*PaystackService, error) {
	if cfg.PaystackSecretKey == "" {
		return nil, fmt.Errorf("Paystack secret key is not configured")
	}
	// The second argument to paystack.NewClient can be a custom http.Client, nil for default.
	client := paystack.NewClient(cfg.PaystackSecretKey, nil)
	log.Println("Successfully initialized Paystack client.")
	return &PaystackService{Client: client, Cfg: cfg}, nil
}

// InitializePayment initializes a payment transaction with Paystack and returns raw response
func (s *PaystackService) InitializePayment(req models.PaystackInitializeRequest, userID string, callbackURL string) (interface{}, error) {
    metadata := map[string]interface{}{
        "user_id": userID,
        "custom_fields": []map[string]string{
            {"display_name": "User ID", "variable_name": "user_id", "value": userID},
            {"display_name": "Purpose", "variable_name": "purpose", "value": "Datacredit Purchase"},
        },
    }

    transactionReq := &paystack.TransactionRequest{
        Email:       req.Email,
        Amount:      float32(req.Amount),
        Currency:    "NGN",
        Metadata:    metadata,
        CallbackURL: callbackURL,
    }

    // Return the raw response for inspection
    resp, err := s.Client.Transaction.Initialize(transactionReq)
    if err != nil {
        return nil, fmt.Errorf("error initializing Paystack transaction: %w", err)
    }

    return resp, nil
}

// VerifyPayment verifies a payment transaction with Paystack using the transaction reference.
func (s *PaystackService) VerifyPayment(reference string) (*paystack.Transaction, error) {
	if reference == "" {
		return nil, fmt.Errorf("payment reference cannot be empty for verification")
	}
	transaction, err := s.Client.Transaction.Verify(reference)
	if err != nil {
		return nil, fmt.Errorf("error verifying Paystack transaction %s: %w", reference, err)
	}
	// Note: `transaction.Status` inside the returned object will indicate "success", "failed", "abandoned"
	return transaction, nil
}

// VerifyWebhookSignature verifies the signature of an incoming Paystack webhook.
// This is a CRITICAL security step.
func (s *PaystackService) VerifyWebhookSignature(requestBody []byte, signatureFromHeader string) bool {
	// IMPORTANT: This is a placeholder implementation.
	// You MUST implement this based on Paystack's documentation:
	// "Verify the event by generating a hash using the same secret key and comparing it to the value in the X-Paystack-Signature header."
	// The hash is typically an HMAC-SHA512 of the request body using your Paystack Secret Key.

	if s.Cfg.PaystackSecretKey == "" {
		log.Println("CRITICAL SECURITY: Paystack secret key not set. Cannot verify webhook signature.")
		return false // Cannot verify without the key
	}
	if signatureFromHeader == "" {
		log.Println("Webhook signature missing from header.")
		return false
	}

	mac := hmac.New(sha512.New, []byte(s.Cfg.PaystackSecretKey))
	_, err := mac.Write(requestBody)
	if err != nil {
		// This error is unlikely for hmac.Write but good to be aware of
		log.Printf("Error writing request body to HMAC: %v", err)
		return false
	}
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(signatureFromHeader), []byte(expectedSignature)) {
		return true
	}

	log.Printf("Webhook signature mismatch. Expected: %s, Got: %s", expectedSignature, signatureFromHeader)
	return false
}

// ProcessSuccessfulPayment is called after a payment is verified (e.g., via webhook).
// It involves updating user datacredit balance in Supabase.
func (s *PaystackService) ProcessSuccessfulPayment(transactionData models.PaystackTransactionData, supabaseService *SupabaseService) error {
	log.Printf("Processing successful Paystack payment for reference: %s, Email: %s, Amount: %d kobo",
		transactionData.Reference, transactionData.Customer.Email, transactionData.Amount)

	var userID string
	// Option 1: Get userID from Paystack metadata (best if you set it reliably)
	if metadataUserID, ok := transactionData.Metadata["user_id"].(string); ok && metadataUserID != "" {
		userID = metadataUserID
	} else {
		// Option 2: Fallback to find user by email (ensure email is unique in profiles table)
		// This assumes a user profile already exists with this email.
		log.Printf("user_id not found in Paystack metadata for ref: %s. Attempting fallback to email: %s", transactionData.Reference, transactionData.Customer.Email)
		profile, err := supabaseService.GetUserByEmail(transactionData.Customer.Email)
		if err != nil {
			return fmt.Errorf("could not find user by email '%s' to credit after payment (Paystack ref: %s): %w",
				transactionData.Customer.Email, transactionData.Reference, err)
		}
		userID = profile.ID
	}

	if userID == "" {
		return fmt.Errorf("unable to determine user ID for crediting from Paystack transaction %s", transactionData.Reference)
	}

	// Amount from Paystack is already in kobo, which is our unit for datacredit_balance.
	datacreditToAdd := int64(transactionData.Amount)
	opDescription := fmt.Sprintf("Datacredit purchase via Paystack (Ref: %s)", transactionData.Reference)
	externalRef := transactionData.Reference

	_, err := supabaseService.UpdateDatacreditBalance(userID, datacreditToAdd, opDescription, &externalRef)
	if err != nil {
		// This is a critical error. Payment received but crediting failed.
		// Implement alerting or a retry mechanism with idempotency.
		log.Printf("CRITICAL ERROR: Paystack payment %s received for user %s, but failed to update datacredit balance: %v",
			transactionData.Reference, userID, err)
		return fmt.Errorf("failed to update user datacredit balance for UserID %s after payment %s: %w",
			userID, transactionData.Reference, err)
	}

	log.Printf("Successfully credited %d datacredit (kobo) to UserID %s for Paystack Ref: %s",
		datacreditToAdd, userID, transactionData.Reference)
	return nil
}

// InitiateWithdrawal processes a withdrawal request by initiating a transfer via Paystack.
func (s *PaystackService) InitiateWithdrawal(req models.WithdrawalRequest, supabaseService *SupabaseService) (string, error) {
	// Amount in req.Amount is datacredit (kobo) to withdraw
	datacreditToWithdrawKobo := req.Amount

	// 1. Check user's datacredit balance from Supabase
	wallet, err := supabaseService.GetOrCreateWallet(req.UserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user wallet for withdrawal: %w", err)
	}

	if wallet.DatacreditBalance < datacreditToWithdrawKobo {
		return "", fmt.Errorf("insufficient datacredit balance for withdrawal. Has: %d kobo, Wants: %d kobo", wallet.DatacreditBalance, datacreditToWithdrawKobo)
	}

	// 2. Create Transfer Recipient with Paystack (or retrieve existing one)
	// This part needs your actual logic for managing recipients.
	// For example, you might fetch bank details from the user's profile or a dedicated table.
	// Let's assume you have a function to get/create a recipient code.
	// For this example, we'll use a placeholder.
	//
	// Example:
	// recipientDetails := &paystack.TransferRecipientRequest{
	// 	Type: "nuban",
	// 	Name: "User Full Name from Profile", // Fetch this
	// 	AccountNumber: "USER_BANK_ACCOUNT_NUMBER", // Fetch this
	// 	BankCode: "USER_BANK_CODE", // Fetch this (Paystack has API to list banks)
	// 	Currency: "NGN",
	// }
	// recipient, err := s.Client.Transfer.CreateRecipient(recipientDetails)
	// if err != nil || !recipient.Status {
	// 	return "", fmt.Errorf("failed to create Paystack transfer recipient: %v - Message: %s", err, recipient.Message)
	// }
	// recipientCode := recipient.Data.RecipientCode
	//
	// For now, using a placeholder:
	recipientCode := "RCP_PLACEHOLDER_RECIPIENT_CODE" // !! REPLACE WITH ACTUAL LOGIC !!
	if recipientCode == "RCP_PLACEHOLDER_RECIPIENT_CODE" {
		log.Println("WARNING: Using placeholder Paystack recipient code for withdrawal.")
		// In a real app, you'd return an error or have robust recipient management.
		// return "", fmt.Errorf("recipient code management not implemented")
	}

	// 3. Initiate Transfer with Paystack
	transferReq := &paystack.TransferRequest{
		Source:    "balance",                         // Transfer from your Paystack balance
		Amount:    float32(datacreditToWithdrawKobo), // Amount in Kobo - corrected to float32
		Recipient: recipientCode,
		Reason:    fmt.Sprintf("Datacredit withdrawal for UserID %s", req.UserID),
		Currency:  "NGN",
		// Reference: You can generate a unique reference for this transfer for idempotency
	}

	log.Printf("Attempting Paystack transfer: UserID %s, Amount %d kobo, Recipient %s", req.UserID, datacreditToWithdrawKobo, recipientCode)
	transferResponse, err := s.Client.Transfer.Initiate(transferReq)
	if err != nil {
		// This could be a network error or an error from Paystack API itself (e.g., bad request)
		log.Printf("Error response from Paystack transfer initiation: %v", err)
		return "", fmt.Errorf("failed to initiate Paystack transfer: %w", err)
	}

	// Check if the transfer initiation was successful
	// The transfer response should contain transfer details
	if transferResponse == nil {
		return "", fmt.Errorf("received nil response from Paystack transfer initiation")
	}

	// Extract transfer code from the response
	// Note: You may need to adjust this based on the actual structure of the Transfer type
	// from the paystack-go library. Check the library documentation for the correct field names.
	transferCode := transferResponse.TransferCode
	if transferCode == "" {
		return "", fmt.Errorf("could not extract transfer_code from Paystack response")
	}

	log.Printf("Paystack transfer successfully initiated. Transfer Code: %s", transferCode)

	// 4. If Paystack transfer initiation is successful, debit user's app credits in Supabase.
	// IMPORTANT: The most robust approach is to wait for a 'transfer.success' webhook from Paystack
	// before debiting the user's balance in your system. This prevents issues if the transfer
	// is initiated but fails later for reasons Paystack can only determine asynchronously.
	//
	// For this example, we'll debit immediately after successful *initiation*.
	// Be aware of the risks and implement a reconciliation process if needed.
	opDescription := fmt.Sprintf("Datacredit withdrawal to bank (Paystack Transfer Code: %s)", transferCode)

	_, err = supabaseService.UpdateDatacreditBalance(req.UserID, -datacreditToWithdrawKobo, opDescription, &transferCode) // Negative amount
	if err != nil {
		log.Printf("CRITICAL ERROR: Initiated Paystack transfer %s but failed to debit datacredit for UserID %s: %v", transferCode, req.UserID, err)
		// What to do here? The money might be on its way via Paystack.
		// - Log for manual reconciliation.
		// - Attempt to flag the transaction.
		// - Depending on Paystack's capabilities, see if an initiated transfer can be cancelled (unlikely for most).
		return "", fmt.Errorf("initiated Paystack transfer (%s) but failed to debit user credits. Error: %w", transferCode, err)
	}

	log.Printf("Datacredit debited for UserID %s after Paystack transfer initiation. Transfer Code: %s", req.UserID, transferCode)
	return transferCode, nil
}
