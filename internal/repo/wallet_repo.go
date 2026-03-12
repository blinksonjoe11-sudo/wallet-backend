package repo

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"wallet-api/internal/app"
	"wallet-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepo struct {
	db *app.App
}

func NewWalletRepo(db *app.App) *WalletRepo {
	return &WalletRepo{
		db: db,
	}
}

func GenerateAccountNumber() string {
	rand.Seed(time.Now().UnixNano())
	number := rand.Intn(900000000) + 100000000
	return fmt.Sprintf("96%d", number)
}

func (r *WalletRepo) CreateWallet(w *model.Wallet) error {

	for {
		w.AccountNumber = GenerateAccountNumber()

		err := r.db.DB.Create(&w).Error

		if err == nil {
			return nil
		}

		if !strings.Contains(err.Error(), "duplicate") {
			return fmt.Errorf("error creating wallet: %v", err)
		}
	}

}

func (r *WalletRepo) ExecTx(fn func(tx *gorm.DB) error) error {
	return r.db.DB.Transaction(fn)

}

func (r *WalletRepo) GetUserById(userId uuid.UUID) (*model.Wallet, error) {
	var w model.Wallet

	err := r.db.DB.Where("user_id = ?", userId).First(&w).Error
	return &w, err
}

func (r *WalletRepo) CreateWalletPin(userId uuid.UUID, hash string) error {

	err := r.db.DB.Model(&model.Wallet{}).Where("user_id = ?", userId).Updates(model.Wallet{
		PinHash: hash,
		PinSet:  true,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *WalletRepo) IncrementAttempts(walletId uuid.UUID, attempt int) error {

	err := r.db.DB.Model(&model.Wallet{}).Where("id = ?", walletId).Updates(model.Wallet{
		FailedPinAttempts: attempt,
	}).Error

	if err != nil {
		return err
	}

	return nil
}

func (r *WalletRepo) LockWallet(walletId uuid.UUID, lockUntil time.Time) error {

	return r.db.DB.Model(&model.Wallet{}).Where("id = ?", walletId).Updates(model.Wallet{
		LockedUntil: &lockUntil,
	}).Error

}

func (r *WalletRepo) ResetLockout(walletId uuid.UUID) error {
	return r.db.DB.Model(&model.Wallet{}).Where("id = ?", walletId).Updates(map[string]any{
		"locked_until":        nil,
		"failed_pin_attempts": 0,
	}).Error
}

func (r *WalletRepo) GetWalletByIdWithLock(tx *gorm.DB, walletId uuid.UUID) (*model.Wallet, error) {

	var w model.Wallet

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", walletId).First(&w).Error

	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (r *WalletRepo) UpdateBalance(tx *gorm.DB, walletId uuid.UUID, amount float64) error {

	return tx.Model(&model.Wallet{}).Where("id = ?", walletId).Update("balance", gorm.Expr("balance + ?", amount)).Error

}

func (r *WalletRepo) CreateTransaction(tx *gorm.DB, record model.Transaction) error {
	return tx.Create(&record).Error
}

func (r *WalletRepo) Freeze(userId uuid.UUID, reason string) error {

	var w model.Wallet

	return r.db.DB.Model(&w).Where("id = ?", userId).Updates(model.Wallet{
		IsFrozen:     true,
		FreezeReason: reason,
	}).Error

}

func (r *WalletRepo) ResetDailySpent() error {
	return r.db.DB.Model(&model.Wallet{}).Where("1 = 1").Updates(map[string]any{
		"daily_spent":          0,
		"daily_limit_reset_at": time.Now(),
	}).Error
}

func (r *WalletRepo) History(walletId uuid.UUID) ([]model.Transaction, error) {

	var transactions []model.Transaction

	err := r.db.DB.Model(&model.Transaction{}).Where("from_wallet_id = ? OR to_wallet_id = ?", walletId, walletId).
		Order("created_at DESC").
		Find(&transactions).Error

	if err != nil {
		return nil, err
	}

	return transactions, nil

}

func (r *WalletRepo) Wallet(walletId uuid.UUID) (model.Wallet, error) {

	var wallet model.Wallet

	err := r.db.DB.Model(&model.Wallet{}).
		Where("id = ?", walletId).
		First(&wallet).Error

	if err != nil {
		return model.Wallet{}, err
	}

	return wallet, nil
}
