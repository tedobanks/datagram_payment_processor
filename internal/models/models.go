
package models

import "time"

// Profile matches your 'profiles' table.
type Profile struct {
	ID              string     `json:"id"` //  (UUID from auth.users)
	Name            *string    `json:"name,omitempty"`
	Username        *string    `json:"username,omitempty"`
	Email           *string    `json:"email,omitempty"`
	InviteCode      *string    `json:"invite_code,omitempty"`
	InvitedByUserID *string    `json:"invited_by_user_id,omitempty"`
	ImageURL        *string    `json:"image_url,omitempty"`
	Role            *string    `json:"role,omitempty"` // Assuming 'roletype' is a string representation
	CreatedAt       time.Time  `json:"created_at,omitempty"`
	UpdatedAt       time.Time  `json:"updated_at,omitempty"`
}

// Wallet matches your 'wallets' table.
type Wallet struct {
	UserID            string    `json:"user_id"` // (FK to profiles.id or auth.users.id)
	DatabyteBalance   int64     `json:"databyte_balance"`
	DatacreditBalance int64     `json:"datacredit_balance"` // Represents NGN value in kobo
	CreatedAt         time.Time `json:"created_at,omitempty"`
	UpdatedAt         time.Time `json:"updated_at,omitempty"`
}

// Transaction matches your 'transactions' table.
type Transaction struct {
	ID                   int64                  `json:"id"`
	UserID               string                 `json:"user_id"` // (FK to profiles.id or auth.users.id)
	Amount               int64                  `json:"amount"`  // Amount in kobo for datacredit, or units for databyte
	BalanceBefore        *int64                 `json:"balance_before,omitempty"` // Assuming 'transactionbalance' refers to this
	BalanceAfter         *int64                 `json:"balance_after,omitempty"`  // You'll need logic to populate this
	Operation            string                 `json:"operation"`                 // Assuming 'transactionoperation' (e.g., "credit_purchase", "withdrawal", "databyte_purchase")
	Description          *string                `json:"description,omitempty"`
	ExternalReferenceID  *string                `json:"external_reference_id,omitempty"` // e.g., Paystack reference
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
	TransactionTimestamp time.Time              `json:"transaction_timestamp,omitempty"`
}

// PaystackInitializeRequest remains the same.
type PaystackInitializeRequest struct {
	Email  string `json:"email" binding:"required,email"`
	Amount int    `json:"amount" binding:"required,gt=0"` // Amount in kobo (for datacredit purchase)
}

// PaystackWebhookPayload remains the same.
type PaystackWebhookPayload struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// PaystackTransactionData remains largely the same, ensure it captures what you need.
type PaystackTransactionData struct {
	ID              int64                  `json:"id"`
	Domain          string                 `json:"domain"`
	Status          string                 `json:"status"`
	Reference       string                 `json:"reference"`
	Amount          int                    `json:"amount"` // in kobo
	Message         *string                `json:"message"`
	GatewayResponse string                 `json:"gateway_response"`
	PaidAt          string                 `json:"paid_at"`
	CreatedAt       string                 `json:"created_at"`
	Channel         string                 `json:"channel"`
	Currency        string                 `json:"currency"`
	IPAddress       string                 `json:"ip_address"`
	Metadata        map[string]interface{} `json:"metadata"` // Crucial for your user_id
	Customer        struct {
		Email     string   `json:"email"`
		FirstName *string  `json:"first_name"`
		LastName  *string  `json:"last_name"`
	} `json:"customer"`
}

// WithdrawalRequest now primarily concerns datacredit (NGN value).
type WithdrawalRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Amount int64  `json:"amount" binding:"required,gt=0"` // Amount of datacredit (kobo) to withdraw
	// Paystack recipient details might be needed here or fetched from user's profile
	// BankCode      string `json:"bank_code" binding:"required"`
	// AccountNumber string `json:"account_number" binding:"required"`
}

// DatabytePurchaseRequest could be a model for users buying Databytes using their Datacredits.
type DatabytePurchaseRequest struct {
	UserID         string `json:"user_id" binding:"required"`
	DatabyteAmount int64  `json:"databyte_amount" binding:"required,gt=0"`
}
