package model

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`

	FromWalletID *uuid.UUID `gorm:"type:uuid;index" json:"from_wallet_id"`
	ToWalletID   *uuid.UUID `gorm:"type:uuid;index" json:"to_wallet_id"`

	Amount float64 `gorm:"type:decimal(20,2);not null" json:"amount"`
	Fee    float64 `gorm:"type:decimal(20,2);default:0" json:"fee"`
	Net    float64 `gorm:"type:decimal(20,2);not null" json:"net"` 

	Type string `gorm:"type:varchar(50);not null" json:"type"`

	Status        string `gorm:"type:varchar(50);default:'pending'" json:"status"`
	FailureReason string `gorm:"type:varchar(255)" json:"failure_reason,omitempty"`

	Reference      string `gorm:"type:varchar(100);uniqueIndex;not null" json:"reference"` 
	IdempotencyKey string `gorm:"type:varchar(255);uniqueIndex" json:"idempotency_key"`    

	//ReversalOfID *uuid.UUID `gorm:"type:uuid" json:"reversal_of_id,omitempty"`

	SenderBalanceAfter   *float64 `gorm:"type:decimal(20,2)" json:"sender_balance_after,omitempty"`
	ReceiverBalanceAfter *float64 `gorm:"type:decimal(20,2)" json:"receiver_balance_after,omitempty"`

	Narration string `db:"narration" json:"narration"`

	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`

	// relationships
	FromWallet *Wallet `gorm:"foreignKey:FromWalletID" json:"-"`
	ToWallet   *Wallet `gorm:"foreignKey:ToWalletID" json:"-"`
	// ReversalOf *Transaction `gorm:"foreignKey:ReversalOfID" json:"-"`
}


const (
	TypeDeposit = "deposit"
	TypeTransfer = "transfer"
	TypeWithdraw = "withdraw"
)

const (
	TypeSuccess = "success"
	TypeFailed = "failed"
)