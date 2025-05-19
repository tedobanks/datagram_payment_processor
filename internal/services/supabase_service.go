// package services

// import (
// 	"fmt"
// 	"log"
// 	"math"
// 	"time"

// 	// your_module_name/config (replace your_module_name)
// 	"github.com/supabase-community/supabase-go"
// 	"github.com/tedobanks/datagram_payment_processor/internal/config"
// 	"github.com/tedobanks/datagram_payment_processor/internal/models"
// 	// supabasestorageuploader "github.com/nedpals/supabase-go/storage"
// )

// // SupabaseService provides methods to interact with Supabase
// // SupabaseService provides methods to interact with Supabase
// type SupabaseService struct {
// 	Client *supabase.Client
// }

// // NewSupabaseService creates a new SupabaseService
// func NewSupabaseService(cfg *config.Config) (*SupabaseService, error) {
// 	client, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to initialize Supabase client: %w", err)
// 	}
// 	log.Println("Successfully connected to Supabase!")
// 	return &SupabaseService{Client: client}, nil
// }

// // GetOrCreateWallet retrieves a user's wallet or creates one if it doesn't exist.
// func (s *SupabaseService) GetOrCreateWallet(userID string) (*models.Wallet, error) {
// 	var wallets []models.Wallet
// 	err := s.Client.DB.From("wallets").Select("*").Eq("user_id", userID).Execute(&wallets)
// 	if err != nil {
// 		return nil, fmt.Errorf("error fetching wallet for user %s: %w", userID, err)
// 	}

// 	if len(wallets) > 0 {
// 		return &wallets[0], nil
// 	}

// 	// Wallet not found, create a new one with default zero balances
// 	newWallet := models.Wallet{
// 		UserID:            userID,
// 		DatabyteBalance:   0,
// 		DatacreditBalance: 0,
// 	}
// 	var createdWallets []models.Wallet
// 	err = s.Client.DB.From("wallets").Insert(newWallet).Execute(&createdWallets)
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating wallet for user %s: %w", userID, err)
// 	}
// 	if len(createdWallets) == 0 {
// 		return nil, fmt.Errorf("failed to create wallet for user %s, no data returned", userID)
// 	}
// 	return &createdWallets[0], nil
// }

// // UpdateDatacreditBalance updates the datacredit_balance for a user.
// // amountKobo can be positive (credit) or negative (debit).
// // Returns the updated wallet.
// func (s *SupabaseService) UpdateDatacreditBalance(userID string, amountKobo int64, operationDescription string, externalRef *string) (*models.Wallet, error) {
// 	wallet, err := s.GetOrCreateWallet(userID)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not get/create wallet for datacredit update: %w", err)
// 	}

// 	balanceBefore := wallet.DatacreditBalance
// 	newBalance := wallet.DatacreditBalance + amountKobo

// 	if newBalance < 0 {
// 		return nil, fmt.Errorf("insufficient datacredit balance for user %s. Has: %d, Tried to change by: %d", userID, wallet.DatacreditBalance, amountKobo)
// 	}

// 	updateData := map[string]interface{}{"datacredit_balance": newBalance}
// 	var updatedWallets []models.Wallet

// 	// Database transaction start (conceptual, actual implementation depends on how you handle atomicity with Supabase from Go)
// 	// For Supabase, you often rely on RLS or call PostgreSQL functions for atomic operations.
// 	// Here, we'll do a read then write. For high concurrency, consider an RPC function.

// 	err = s.Client.DB.From("wallets").Update(updateData).Eq("user_id", userID).Execute(&updatedWallets)
// 	if err != nil {
// 		return nil, fmt.Errorf("error updating datacredit balance in Supabase for user %s: %w", userID, err)
// 	}
// 	if len(updatedWallets) == 0 {
// 		return nil, fmt.Errorf("user wallet %s not found for datacredit update or no change made", userID)
// 	}

// 	// Log transaction
// 	transaction := models.Transaction{
// 		UserID:              userID,
// 		Amount:              amountKobo, // This is the change amount
// 		BalanceBefore:       &balanceBefore,
// 		BalanceAfter:        &newBalance,
// 		Operation:           "datacredit_update", // More specific e.g., "datacredit_purchase", "datacredit_withdrawal"
// 		Description:         &operationDescription,
// 		ExternalReferenceID: externalRef,
// 		// Metadata: // add if any relevant metadata for this specific transaction
// 	}
// 	if err := s.LogTransaction(transaction); err != nil {
// 		log.Printf("WARNING: Datacredit balance updated for user %s, but failed to log transaction: %v", userID, err)
// 		// Decide on error handling here: is it critical if logging fails?
// 	}

// 	return &updatedWallets[0], nil
// }

// // UpdateDatabyteBalance updates the databyte_balance for a user.
// // amountDatabyte can be positive (credit) or negative (debit).
// func (s *SupabaseService) UpdateDatabyteBalance(userID string, amountDatabyte int64, operationDescription string, externalRef *string) (*models.Wallet, error) {
// 	wallet, err := s.GetOrCreateWallet(userID)
// 	if err != nil {
// 		return nil, fmt.Errorf("could not get/create wallet for databyte update: %w", err)
// 	}

// 	balanceBefore := wallet.DatabyteBalance
// 	newBalance := wallet.DatabyteBalance + amountDatabyte

// 	if newBalance < 0 {
// 		return nil, fmt.Errorf("insufficient databyte balance for user %s. Has: %d, Tried to change by: %d", userID, wallet.DatabyteBalance, amountDatabyte)
// 	}

// 	updateData := map[string]interface{}{"databyte_balance": newBalance}
// 	var updatedWallets []models.Wallet
// 	err = s.Client.DB.From("wallets").Update(updateData).Eq("user_id", userID).Execute(&updatedWallets)
// 	if err != nil {
// 		return nil, fmt.Errorf("error updating databyte balance for user %s: %w", userID, err)
// 	}
// 	if len(updatedWallets) == 0 {
// 		return nil, fmt.Errorf("user wallet %s not found for databyte update", userID)
// 	}

// 	// Log transaction
// 	transaction := models.Transaction{
// 		UserID:              userID,
// 		Amount:              amountDatabyte,
// 		BalanceBefore:       &balanceBefore,
// 		BalanceAfter:        &newBalance,
// 		Operation:           "databyte_update", // e.g., "databyte_purchase_with_credit", "databyte_usage"
// 		Description:         &operationDescription,
// 		ExternalReferenceID: externalRef,
// 	}
// 	if err := s.LogTransaction(transaction); err != nil {
// 		log.Printf("WARNING: Databyte balance updated for user %s, but failed to log transaction: %v", userID, err)
// 	}
// 	return &updatedWallets[0], nil
// }

// // PurchaseDatabytesWithDatacredit converts datacredit to databytes for a user.
// func (s *SupabaseService) PurchaseDatabytesWithDatacredit(userID string, databyteAmountToPurchase int64) (*models.Wallet, error) {
// 	if databyteAmountToPurchase <= 0 {
// 		return nil, fmt.Errorf("databyte amount to purchase must be positive")
// 	}

// 	// 1. Calculate the "exact" cost in kobo as a float
// 	exactKoboCostFloat := float64(databyteAmountToPurchase) / float64(config.DATABYTES_PER_DATACREDIT_KOBO)

// 	// 2. Determine the actual kobo to debit by rounding up to the nearest whole kobo.
// 	// This ensures you always debit an integer amount of kobo.
// 	actualKoboToDebit := int64(math.Ceil(exactKoboCostFloat))

// 	// 3. Enforce a minimum charge of 1 kobo if any databytes are being purchased
// 	// (i.e., if exactKoboCostFloat is > 0 but was rounded down to 0 by int64 conversion before ceil,
// 	// or if exactKoboCostFloat was < 1 but > 0, math.Ceil makes it 1 anyway.
// 	// This just makes sure we charge at least 1 kobo if any cost is incurred.)
// 	if actualKoboToDebit == 0 && databyteAmountToPurchase > 0 {
// 		actualKoboToDebit = 1 // Minimum charge of 1 kobo
// 	}

// 	datacreditCostKobo := actualKoboToDebit

// 	wallet, err := s.GetOrCreateWallet(userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Check if user has enough datacredit
// 	if wallet.DatacreditBalance < datacreditCostKobo {
// 		return nil, fmt.Errorf("insufficient datacredit balance. Required: %d kobo, Available: %d kobo", datacreditCostKobo, wallet.DatacreditBalance)
// 	}

// 	// Perform updates (ideally within a database transaction or RPC call for atomicity)
// 	// 1. Debit Datacredit
// 	newDatacreditBalance := wallet.DatacreditBalance - datacreditCostKobo
// 	datacreditBalanceBefore := wallet.DatacreditBalance

// 	// 2. Credit Databyte
// 	newDatabyteBalance := wallet.DatabyteBalance + databyteAmountToPurchase
// 	databyteBalanceBefore := wallet.DatabyteBalance

// 	// Update wallet in DB
// 	// For true atomicity, this should be a single operation or an RPC call.
// 	// Here, we'll do two separate updates for simplicity of example, but this is NOT atomic.
// 	// A better way is to update both balances in one call if the DB schema/PostgREST allows,
// 	// or use a plpgsql function.

// 	// Update datacredit
// 	var updatedWallets []models.Wallet
// 	err = s.Client.DB.From("wallets").Update(map[string]interface{}{"datacredit_balance": newDatacreditBalance}).Eq("user_id", userID).Execute(&updatedWallets)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to debit datacredit: %w", err)
// 	}
// 	if len(updatedWallets) == 0 {
// 		return nil, fmt.Errorf("wallet not found during datacredit debit")
// 	}

// 	// Update databyte
// 	err = s.Client.DB.From("wallets").Update(map[string]interface{}{"databyte_balance": newDatabyteBalance}).Eq("user_id", userID).Execute(&updatedWallets)
// 	if err != nil {
// 		// Attempt to rollback or flag inconsistency if possible
// 		log.Printf("CRITICAL: Failed to credit databyte after debiting datacredit for user %s. Manual intervention likely needed.", userID)
// 		return nil, fmt.Errorf("failed to credit databyte: %w", err)
// 	}
// 	if len(updatedWallets) == 0 {
// 		return nil, fmt.Errorf("wallet not found during databyte credit")
// 	}

// 	// Log the two-part transaction
// 	descDebit := fmt.Sprintf("Purchase of %d databytes", databyteAmountToPurchase)
// 	s.LogTransaction(models.Transaction{
// 		UserID:        userID,
// 		Amount:        -datacreditCostKobo, // Negative for debit
// 		BalanceBefore: &datacreditBalanceBefore,
// 		BalanceAfter:  &newDatacreditBalance,
// 		Operation:     "datacredit_debit_for_databyte",
// 		Description:   &descDebit,
// 	})

// 	descCredit := fmt.Sprintf("Purchased with %d datacredit (kobo)", datacreditCostKobo)
// 	s.LogTransaction(models.Transaction{
// 		UserID:        userID,
// 		Amount:        databyteAmountToPurchase, // Positive for credit
// 		BalanceBefore: &databyteBalanceBefore,
// 		BalanceAfter:  &newDatabyteBalance,
// 		Operation:     "databyte_credit_from_purchase",
// 		Description:   &descCredit,
// 	})

// 	return &updatedWallets[0], nil
// }

// // LogTransaction records a financial or currency operation.
// func (s *SupabaseService) LogTransaction(tx models.Transaction) error {
// 	// Ensure TransactionTimestamp is set if not already
// 	if tx.TransactionTimestamp.IsZero() {
// 		tx.TransactionTimestamp = time.Now()
// 	}

// 	var results []models.Transaction
// 	err := s.Client.DB.From("transactions").Insert(tx).Execute(&results)
// 	if err != nil {
// 		return fmt.Errorf("error logging transaction for user %s: %w", tx.UserID, err)
// 	}
// 	if len(results) == 0 {
// 		return fmt.Errorf("failed to log transaction, no data returned")
// 	}
// 	log.Printf("Transaction logged successfully for user %s, operation: %s", tx.UserID, tx.Operation)
// 	return nil
// }

// // GetUserProfile retrieves a user's profile.
// func (s *SupabaseService) GetUserProfile(userID string) (*models.Profile, error) {
// 	var profiles []models.Profile
// 	err := s.Client.DB.From("profiles").Select("*").Eq("id", userID).Execute(&profiles)
// 	if err != nil {
// 		return nil, fmt.Errorf("error fetching profile for user %s: %w", userID, err)
// 	}
// 	if len(profiles) == 0 {
// 		return nil, fmt.Errorf("profile not found for user %s", userID)
// 	}
// 	return &profiles[0], nil
// }

// // GetUserByEmail retrieves a user's profile by email.
// func (s *SupabaseService) GetUserByEmail(email string) (*models.Profile, error) {
// 	var profiles []models.Profile
// 	// Ensure your 'profiles' table has a unique constraint on email or handle multiple results appropriately
// 	err := s.Client.DB.From("profiles").Select("id, email").Eq("email", email).Limit(1).Execute(&profiles)
// 	if err != nil {
// 		return nil, fmt.Errorf("error fetching profile by email %s: %w", email, err)
// 	}
// 	if len(profiles) == 0 {
// 		return nil, fmt.Errorf("profile not found for email %s", email)
// 	}
// 	return &profiles[0], nil
// }


package services

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/supabase-community/postgrest-go"
	"github.com/supabase-community/supabase-go"
	"github.com/tedobanks/datagram_payment_processor/internal/config"
	"github.com/tedobanks/datagram_payment_processor/internal/models"
)

// SupabaseService provides methods to interact with Supabase
type SupabaseService struct {
	Client *supabase.Client
	DB     *postgrest.Client
}

// NewSupabaseService creates a new SupabaseService
func NewSupabaseService(cfg *config.Config) (*SupabaseService, error) {
	client, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", err)
	}
	log.Println("Successfully connected to Supabase!")
	
	// Create PostgREST client directly
	db := postgrest.NewClient(cfg.SupabaseURL+"/rest/v1", cfg.SupabaseServiceKey, nil)
	
	return &SupabaseService{
		Client: client,
		DB:     db,
	}, nil
}

// GetOrCreateWallet retrieves a user's wallet or creates one if it doesn't exist.
func (s *SupabaseService) GetOrCreateWallet(userID string) (*models.Wallet, error) {
	data, _, err := s.DB.From("wallets").Select("*", "", false).Eq("user_id", userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error fetching wallet for user %s: %w", userID, err)
	}

	var wallets []models.Wallet
	if err := json.Unmarshal(data, &wallets); err != nil {
		return nil, fmt.Errorf("error unmarshaling wallet data for user %s: %w", userID, err)
	}

	if len(wallets) > 0 {
		return &wallets[0], nil
	}

	// Wallet not found, create a new one with default zero balances
	newWallet := models.Wallet{
		UserID:            userID,
		DatabyteBalance:   0,
		DatacreditBalance: 0,
	}
	
	data, _, err = s.DB.From("wallets").Insert(newWallet, false, "", "", "").Execute()
	if err != nil {
		return nil, fmt.Errorf("error creating wallet for user %s: %w", userID, err)
	}
	
	var createdWallets []models.Wallet
	if err := json.Unmarshal(data, &createdWallets); err != nil {
		return nil, fmt.Errorf("error unmarshaling created wallet data for user %s: %w", userID, err)
	}
	
	if len(createdWallets) == 0 {
		return nil, fmt.Errorf("failed to create wallet for user %s, no data returned", userID)
	}
	return &createdWallets[0], nil
}

// UpdateDatacreditBalance updates the datacredit_balance for a user.
// amountKobo can be positive (credit) or negative (debit).
// Returns the updated wallet.
func (s *SupabaseService) UpdateDatacreditBalance(userID string, amountKobo int64, operationDescription string, externalRef *string) (*models.Wallet, error) {
	wallet, err := s.GetOrCreateWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("could not get/create wallet for datacredit update: %w", err)
	}

	balanceBefore := wallet.DatacreditBalance
	newBalance := wallet.DatacreditBalance + amountKobo

	if newBalance < 0 {
		return nil, fmt.Errorf("insufficient datacredit balance for user %s. Has: %d, Tried to change by: %d", userID, wallet.DatacreditBalance, amountKobo)
	}

	updateData := models.Wallet{
		DatacreditBalance: newBalance,
	}
	
	data, _, err := s.DB.From("wallets").Update(updateData, "", "").Eq("user_id", userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error updating datacredit balance in Supabase for user %s: %w", userID, err)
	}
	
	var updatedWallets []models.Wallet
	if err := json.Unmarshal(data, &updatedWallets); err != nil {
		return nil, fmt.Errorf("error unmarshaling updated wallet data for user %s: %w", userID, err)
	}
	
	if len(updatedWallets) == 0 {
		return nil, fmt.Errorf("user wallet %s not found for datacredit update or no change made", userID)
	}

	// Log transaction
	transaction := models.Transaction{
		UserID:              userID,
		Amount:              amountKobo,
		BalanceBefore:       &balanceBefore,
		BalanceAfter:        &newBalance,
		Operation:           "datacredit_update",
		Description:         &operationDescription,
		ExternalReferenceID: externalRef,
		TransactionTimestamp: time.Now(),
	}
	if err := s.LogTransaction(transaction); err != nil {
		log.Printf("WARNING: Datacredit balance updated for user %s, but failed to log transaction: %v", userID, err)
	}

	return &updatedWallets[0], nil
}

// UpdateDatabyteBalance updates the databyte_balance for a user.
// amountDatabyte can be positive (credit) or negative (debit).
func (s *SupabaseService) UpdateDatabyteBalance(userID string, amountDatabyte int64, operationDescription string, externalRef *string) (*models.Wallet, error) {
	wallet, err := s.GetOrCreateWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("could not get/create wallet for databyte update: %w", err)
	}

	balanceBefore := wallet.DatabyteBalance
	newBalance := wallet.DatabyteBalance + amountDatabyte

	if newBalance < 0 {
		return nil, fmt.Errorf("insufficient databyte balance for user %s. Has: %d, Tried to change by: %d", userID, wallet.DatabyteBalance, amountDatabyte)
	}

	updateData := models.Wallet{
		DatabyteBalance: newBalance,
	}
	
	data, _, err := s.DB.From("wallets").Update(updateData, "", "").Eq("user_id", userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error updating databyte balance for user %s: %w", userID, err)
	}
	
	var updatedWallets []models.Wallet
	if err := json.Unmarshal(data, &updatedWallets); err != nil {
		return nil, fmt.Errorf("error unmarshaling updated wallet data for user %s: %w", userID, err)
	}
	
	if len(updatedWallets) == 0 {
		return nil, fmt.Errorf("user wallet %s not found for databyte update", userID)
	}

	// Log transaction
	transaction := models.Transaction{
		UserID:              userID,
		Amount:              amountDatabyte,
		BalanceBefore:       &balanceBefore,
		BalanceAfter:        &newBalance,
		Operation:           "databyte_update",
		Description:         &operationDescription,
		ExternalReferenceID: externalRef,
		TransactionTimestamp: time.Now(),
	}
	if err := s.LogTransaction(transaction); err != nil {
		log.Printf("WARNING: Databyte balance updated for user %s, but failed to log transaction: %v", userID, err)
	}
	return &updatedWallets[0], nil
}

// PurchaseDatabytesWithDatacredit converts datacredit to databytes for a user.
func (s *SupabaseService) PurchaseDatabytesWithDatacredit(userID string, databyteAmountToPurchase int64) (*models.Wallet, error) {
	if databyteAmountToPurchase <= 0 {
		return nil, fmt.Errorf("databyte amount to purchase must be positive")
	}

	// Calculate the exact cost in kobo
	exactKoboCostFloat := float64(databyteAmountToPurchase) / config.DATABYTES_PER_DATACREDIT_KOBO
	
	// Round up to ensure we charge at least the minimum required
	actualKoboToDebit := int64(math.Ceil(exactKoboCostFloat))
	
	// Enforce minimum charge of 1 kobo
	if actualKoboToDebit <= 0 && databyteAmountToPurchase > 0 {
		actualKoboToDebit = 1
	}

	wallet, err := s.GetOrCreateWallet(userID)
	if err != nil {
		return nil, err
	}

	// Check if user has enough datacredit
	if wallet.DatacreditBalance < actualKoboToDebit {
		return nil, fmt.Errorf("insufficient datacredit balance. Required: %d kobo, Available: %d kobo", actualKoboToDebit, wallet.DatacreditBalance)
	}

	// Store original balances for transaction logging
	datacreditBalanceBefore := wallet.DatacreditBalance
	databyteBalanceBefore := wallet.DatabyteBalance
	
	// Calculate new balances
	newDatacreditBalance := wallet.DatacreditBalance - actualKoboToDebit
	newDatabyteBalance := wallet.DatabyteBalance + databyteAmountToPurchase

	// Update both balances in a single operation for better atomicity
	updateData := models.Wallet{
		DatacreditBalance: newDatacreditBalance,
		DatabyteBalance:   newDatabyteBalance,
		UserID:            userID, // Include UserID to ensure proper update
	}
	
	data, _, err := s.DB.From("wallets").Update(updateData, "", "").Eq("user_id", userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balances: %w", err)
	}
	
	var updatedWallets []models.Wallet
	if err := json.Unmarshal(data, &updatedWallets); err != nil {
		return nil, fmt.Errorf("error unmarshaling updated wallet data: %w", err)
	}
	
	if len(updatedWallets) == 0 {
		return nil, fmt.Errorf("wallet not found during balance update")
	}

	// Log the datacredit debit transaction
	descDebit := fmt.Sprintf("Purchase of %d databytes", databyteAmountToPurchase)
	if err := s.LogTransaction(models.Transaction{
		UserID:               userID,
		Amount:               -actualKoboToDebit, // Negative for debit
		BalanceBefore:        &datacreditBalanceBefore,
		BalanceAfter:         &newDatacreditBalance,
		Operation:            "datacredit_debit_for_databyte",
		Description:          &descDebit,
		TransactionTimestamp: time.Now(),
	}); err != nil {
		log.Printf("WARNING: Failed to log datacredit debit transaction for user %s: %v", userID, err)
	}

	// Log the databyte credit transaction
	descCredit := fmt.Sprintf("Purchased with %d datacredit (kobo)", actualKoboToDebit)
	if err := s.LogTransaction(models.Transaction{
		UserID:               userID,
		Amount:               databyteAmountToPurchase, // Positive for credit
		BalanceBefore:        &databyteBalanceBefore,
		BalanceAfter:         &newDatabyteBalance,
		Operation:            "databyte_credit_from_purchase",
		Description:          &descCredit,
		TransactionTimestamp: time.Now(),
	}); err != nil {
		log.Printf("WARNING: Failed to log databyte credit transaction for user %s: %v", userID, err)
	}

	return &updatedWallets[0], nil
}

// LogTransaction records a financial or currency operation.
func (s *SupabaseService) LogTransaction(tx models.Transaction) error {
	// Ensure TransactionTimestamp is set if not already
	if tx.TransactionTimestamp.IsZero() {
		tx.TransactionTimestamp = time.Now()
	}

	data, _, err := s.DB.From("transactions").Insert(tx, false, "", "", "").Execute()
	if err != nil {
		return fmt.Errorf("error logging transaction for user %s: %w", tx.UserID, err)
	}
	
	var results []models.Transaction
	if err := json.Unmarshal(data, &results); err != nil {
		return fmt.Errorf("error unmarshaling transaction result for user %s: %w", tx.UserID, err)
	}
	
	if len(results) == 0 {
		return fmt.Errorf("failed to log transaction, no data returned")
	}
	log.Printf("Transaction logged successfully for user %s, operation: %s", tx.UserID, tx.Operation)
	return nil
}

// GetUserProfile retrieves a user's profile.
func (s *SupabaseService) GetUserProfile(userID string) (*models.Profile, error) {
	data, _, err := s.DB.From("profiles").Select("*", "", false).Eq("id", userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("error fetching profile for user %s: %w", userID, err)
	}
	
	var profiles []models.Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("error unmarshaling profile data for user %s: %w", userID, err)
	}
	
	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found for user %s", userID)
	}
	return &profiles[0], nil
}

// GetUserByEmail retrieves a user's profile by email.
func (s *SupabaseService) GetUserByEmail(email string) (*models.Profile, error) {
	data, _, err := s.DB.From("profiles").Select("id, email", "", false).Eq("email", email).Limit(1, "").Execute()
	if err != nil {
		return nil, fmt.Errorf("error fetching profile by email %s: %w", email, err)
	}
	
	var profiles []models.Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, fmt.Errorf("error unmarshaling profile data for email %s: %w", email, err)
	}
	
	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found for email %s", email)
	}
	return &profiles[0], nil
}