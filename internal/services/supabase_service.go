
package services

import (
	// "encoding/json"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/supabase-community/supabase-go"
	"github.com/tedobanks/datagram_payment_processor/internal/config"
	"github.com/tedobanks/datagram_payment_processor/internal/models"
)

// SupabaseService provides methods to interact with Supabase
type SupabaseService struct {
	Client *supabase.Client

}

func NewSupabaseService(cfg *config.Config) (*SupabaseService, error) {
	client, err := supabase.NewClient(cfg.SupabaseURL, cfg.SupabaseServiceKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Supabase client: %w", err)
	}

	// Test connection
	var result []map[string]interface{}
	_, err = client.From("wallets").Select("*", "exact", false).Limit(1, "").ExecuteTo(&result)
	if err != nil {
		return nil, fmt.Errorf("supabase connection test failed: %w", err)
	}

	log.Println("Successfully connected to Supabase!")
	return &SupabaseService{Client: client}, nil
}

func (s *SupabaseService) GetOrCreateWallet(userID string) (*models.Wallet, error) {
	var wallets []models.Wallet
	
	_, err := s.Client.From("wallets").
		Select("*", "exact", false).
		Eq("user_id", userID).
		ExecuteTo(&wallets)
		
	if err != nil {
		return nil, fmt.Errorf("error fetching wallet: %w", err)
	}

	if len(wallets) > 0 {
		return &wallets[0], nil
	}

	// Create new wallet
	newWallet := models.Wallet{
		UserID:            userID,
		DatabyteBalance:   0,
		DatacreditBalance: 0,
	}

	var created []models.Wallet
	_, err = s.Client.From("wallets").
		Insert(newWallet, false, "", "", "").
		ExecuteTo(&created)
		
	if err != nil {
		return nil, fmt.Errorf("error creating wallet: %w", err)
	}

	if len(created) == 0 {
		return nil, fmt.Errorf("no data returned after wallet creation")
	}

	return &created[0], nil
}

func (s *SupabaseService) UpdateDatacreditBalance(userID string, amountKobo int64, operationDescription string, externalRef *string) (*models.Wallet, error) {
	wallet, err := s.GetOrCreateWallet(userID)
	if err != nil {
		return nil, fmt.Errorf("could not get/create wallet: %w", err)
	}

	newBalance := wallet.DatacreditBalance + amountKobo
	if newBalance < 0 {
		return nil, fmt.Errorf("insufficient datacredit balance")
	}

	updateData := map[string]interface{}{
		"datacredit_balance": newBalance,
	}

	var updatedWallets []models.Wallet
	_, err = s.Client.From("wallets").
		Update(updateData, "", "").
		Eq("user_id", userID).
		ExecuteTo(&updatedWallets)
	
	if err != nil {
		return nil, fmt.Errorf("error updating balance: %w", err)
	}

	if len(updatedWallets) == 0 {
		return nil, fmt.Errorf("no wallet updated")
	}

	// Log transaction (implementation omitted for brevity)
	return &updatedWallets[0], nil
}

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

	updateData := map[string]interface{}{
		"databyte_balance": newBalance,
	}

	var updatedWallets []models.Wallet
	_, err = s.Client.From("wallets").
		Update(updateData, "", "").
		Eq("user_id", userID).
		ExecuteTo(&updatedWallets)
	if err != nil {
		return nil, fmt.Errorf("error updating databyte balance for user %s: %w", userID, err)
	}

	if len(updatedWallets) == 0 {
		return nil, fmt.Errorf("user wallet %s not found for databyte update", userID)
	}

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

func (s *SupabaseService) PurchaseDatabytesWithDatacredit(userID string, databyteAmountToPurchase int64) (*models.Wallet, error) {
	if databyteAmountToPurchase <= 0 {
		return nil, fmt.Errorf("databyte amount to purchase must be positive")
	}

	exactKoboCostFloat := float64(databyteAmountToPurchase) / config.DATABYTES_PER_DATACREDIT_KOBO
	actualKoboToDebit := int64(math.Ceil(exactKoboCostFloat))
	if actualKoboToDebit <= 0 {
		actualKoboToDebit = 1
	}

	wallet, err := s.GetOrCreateWallet(userID)
	if err != nil {
		return nil, err
	}

	if wallet.DatacreditBalance < actualKoboToDebit {
		return nil, fmt.Errorf("insufficient datacredit balance. Required: %d kobo, Available: %d kobo", actualKoboToDebit, wallet.DatacreditBalance)
	}

	datacreditBalanceBefore := wallet.DatacreditBalance
	databyteBalanceBefore := wallet.DatabyteBalance

	newDatacreditBalance := wallet.DatacreditBalance - actualKoboToDebit
	newDatabyteBalance := wallet.DatabyteBalance + databyteAmountToPurchase

	updateData := map[string]interface{}{
		"datacredit_balance": newDatacreditBalance,
		"databyte_balance":   newDatabyteBalance,
	}

	var updatedWallets []models.Wallet
	_, err = s.Client.From("wallets").
		Update(updateData, "", "").
		Eq("user_id", userID).
		ExecuteTo(&updatedWallets)
	if err != nil {
		return nil, fmt.Errorf("failed to update wallet balances: %w", err)
	}

	if len(updatedWallets) == 0 {
		return nil, fmt.Errorf("wallet not found during balance update")
	}

	descDebit := fmt.Sprintf("Purchase of %d databytes", databyteAmountToPurchase)
	if err := s.LogTransaction(models.Transaction{
		UserID:               userID,
		Amount:               -actualKoboToDebit,
		BalanceBefore:        &datacreditBalanceBefore,
		BalanceAfter:         &newDatacreditBalance,
		Operation:            "datacredit_debit_for_databyte",
		Description:          &descDebit,
		TransactionTimestamp: time.Now(),
	}); err != nil {
		log.Printf("WARNING: Failed to log datacredit debit transaction for user %s: %v", userID, err)
	}

	descCredit := fmt.Sprintf("Purchased with %d datacredit (kobo)", actualKoboToDebit)
	if err := s.LogTransaction(models.Transaction{
		UserID:               userID,
		Amount:               databyteAmountToPurchase,
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

func (s *SupabaseService) LogTransaction(tx models.Transaction) error {
	if tx.TransactionTimestamp.IsZero() {
		tx.TransactionTimestamp = time.Now()
	}

	var results []models.Transaction
	_, err := s.Client.From("transactions").
		Insert(tx, false, "", "", "").
		ExecuteTo(&results)
	if err != nil {
		return fmt.Errorf("error logging transaction for user %s: %w", tx.UserID, err)
	}

	if len(results) == 0 {
		return fmt.Errorf("failed to log transaction, no data returned")
	}

	log.Printf("Transaction logged successfully for user %s, operation: %s", tx.UserID, tx.Operation)
	return nil
}

func (s *SupabaseService) GetUserProfile(userID string) (*models.Profile, error) {
	var profiles []models.Profile
	_, err := s.Client.From("profiles").
		Select("*", "", false).
		Eq("id", userID).
		ExecuteTo(&profiles)
	if err != nil {
		return nil, fmt.Errorf("error fetching profile for user %s: %w", userID, err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found for user %s", userID)
	}
	return &profiles[0], nil
}

func (s *SupabaseService) GetUserByEmail(email string) (*models.Profile, error) {
	var profiles []models.Profile
	_, err := s.Client.From("profiles").
		Select("id, email", "", false).
		Eq("email", email).
		Limit(1, "").
		ExecuteTo(&profiles)
	if err != nil {
		return nil, fmt.Errorf("error fetching profile by email %s: %w", email, err)
	}

	if len(profiles) == 0 {
		return nil, fmt.Errorf("profile not found for email %s", email)
	}
	return &profiles[0], nil
}
