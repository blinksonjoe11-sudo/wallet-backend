package model

import (
	"time"

	"github.com/google/uuid"
)

type Wallet struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`

	Balance  float64 `gorm:"type:decimal(20,2);default:0;not null" json:"balance"`
	Currency string  `gorm:"type:varchar(10);default:'NGN';not null" json:"currency"`

	AccountNumber string `gorm:"type:varchar(11);uniqueIndex;not null" json:"account_number"`

	PinHash string `gorm:"type:varchar(255)" json:"-"`
	PinSet  bool   `gorm:"default:false" json:"pin_set"`

	DailyLimit        float64   `gorm:"type:decimal(20,2);default:1000000" json:"daily_limit"`
	SingleLimit       float64   `gorm:"type:decimal(20,2);default:500000" json:"single_limit"`
	DailySpent        float64   `gorm:"type:decimal(20,2);default:0" json:"daily_spent"`
	DailyLimitResetAt time.Time `gorm:"not null" json:"daily_limit_reset_at"`

	FailedPinAttempts int `gorm:"default:0"`
	LockedUntil       *time.Time

	IsFrozen     bool   `gorm:"default:false" json:"is_frozen"`
	FreezeReason string `gorm:"type:varchar(255)" json:"freeze_reason,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
