package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
	"wallet-api/internal/model"
	"wallet-api/internal/repo"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletService struct {
	repo *repo.WalletRepo
}

func NewWalletService(repo *repo.WalletRepo) *WalletService {
	return &WalletService{
		repo: repo,
	}
}

func GenerateReference() (string, error) {

	bytes := make([]byte, 16)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func (svc *WalletService) Fund(userId uuid.UUID, amount float64) error {

	if amount <= 0 {
		return errors.New("Amount should be greater than 0")
	}

	// get user by id

	w, err := svc.repo.GetUserById(userId)

	if err != nil {
		return err
	}

	if w.IsFrozen == true {
		return errors.New("wallet frozen")
	}

	if w.LockedUntil != nil && w.LockedUntil.After(time.Now()) {
		return errors.New("wallet is locked, try again later")
	}

	return svc.repo.ExecTx(func(tx *gorm.DB) error {

		locakedWallet, err := svc.repo.GetWalletByIdWithLock(tx, w.ID)
		if err != nil {
			return errors.New("wallet not found")
		}

		// update balance

		if err := svc.repo.UpdateBalance(tx, locakedWallet.ID, amount); err != nil {
			return errors.New("Failed to update balance")
		}

		// record transaction

		newBalance := locakedWallet.Balance + amount
		now := time.Now()
		ref, _ := GenerateReference()
		record := model.Transaction{
			ID:                   uuid.New(),
			ToWalletID:           &locakedWallet.ID,
			Amount:               amount,
			Fee:                  0,
			Net:                  amount,
			Type:                 model.TypeDeposit,
			Status:               model.TypeSuccess,
			Reference:            "TX" + ref,
			ReceiverBalanceAfter: &newBalance,
			IdempotencyKey:       uuid.New().String(),
			Narration:            "wallet funding",
			ProcessedAt:          &now,
		}

		return svc.repo.CreateTransaction(tx, record)

	})

}

func (svc *WalletService) Transfer(pin string, senderId uuid.UUID, accountNumber string, amount float64, narration string) error {

	const (
		fee = 10
	)
	// check user pin

	if pin == "" || accountNumber == "" {
		return errors.New("Fields are required")
	}

	if amount <= 0 {
		return errors.New("Invalid Amount")
	}

	s, err := svc.repo.GetUserById(senderId)

	if err != nil {
		return err
	}

	if s.IsFrozen == true {
		return errors.New("wallet frozen")
	}

	if s.LockedUntil != nil && s.LockedUntil.After(time.Now()) {
		return errors.New("wallet is locked, try again later")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(s.PinHash), []byte(pin)); err != nil {
		return errors.New("Invalid Pin")
	}

	// get reciver balance before

	return svc.repo.ExecTx(func(tx *gorm.DB) error {

		// userId is the senderId
		var sender model.Wallet

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", s.ID).First(&sender).Error; err != nil {
			return errors.New("Couldnt get the user")
		}

		if sender.Balance < amount+fee {
			return errors.New("Insufficient Funds")
		}

		if amount+fee > sender.SingleLimit {
			return errors.New("Amount exceeds single transfer limit")
		}

		if sender.DailySpent+amount+fee > sender.DailyLimit {
			return errors.New("Amount exceeds daily limit")
		}

		// receiver

		var receiver model.Wallet

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("account_number = ?", accountNumber).First(&receiver).Error; err != nil {
			return errors.New("Invalid Account Number")
		}

		sender.Balance -= amount + fee
		sender.DailySpent += amount + fee

		receiver.Balance += amount

		if err := tx.Save(&sender).Error; err != nil {
			return err
		}

		if err := tx.Save(&receiver).Error; err != nil {
			return err
		}

		newSenderBalance := sender.Balance
		newRecieverBal := receiver.Balance

		ref, _ := GenerateReference()
		now := time.Now()
		// create transaction
		record := model.Transaction{
			ID:                   uuid.New(),
			FromWalletID:         &sender.ID,
			ToWalletID:           &receiver.ID,
			Amount:               amount,
			Fee:                  fee,
			Net:                  amount,
			Type:                 model.TypeTransfer,
			Status:               model.TypeSuccess,
			Reference:            "TX" + ref,
			SenderBalanceAfter:   &newSenderBalance,
			ReceiverBalanceAfter: &newRecieverBal,
			IdempotencyKey:       uuid.New().String(),
			Narration:            narration,
			ProcessedAt:          &now,
		}

		return svc.repo.CreateTransaction(tx, record)

	})

}

func (svc *WalletService) CreateWalletPinUser(userId uuid.UUID, pin string) error {

	// check if user alread set pin

	// get user wallet

	if len(pin) != 5 {
		return errors.New("Pin must be 5 characters")
	}

	w, err := svc.repo.GetUserById(userId)

	if err != nil {
		return err
	}

	if w.PinSet == true {
		return errors.New("User already set password")
	}

	hashPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)

	if err != nil {
		return errors.New("Error hashing")
	}

	hash := string(hashPin)

	if err := svc.repo.CreateWalletPin(userId, hash); err != nil {
		return err
	}

	return nil

}

func (svc *WalletService) UpdateWalletPinUser(userId uuid.UUID, oldPin string, newPin string, confirmNewPin string) error {

	if newPin == "" || oldPin == "" {
		return errors.New("Pin must not be empty")
	}

	if len(newPin) != 5 {
		return errors.New("Pin must be 5 characters")
	}

	if newPin != confirmNewPin {
		return errors.New("Pin must be thesame")
	}

	// get and verify old pin
	u, err := svc.repo.GetUserById(userId)

	if err != nil {
		return err
	}

	if u.PinSet == false {
		return errors.New("Create a Pin First")
	}

	// check if wallet is locked
	if u.LockedUntil != nil && u.LockedUntil.After(time.Now()) {
		return errors.New("wallet is locked, try again later")
	}

	// verify old pin from user

	if err := bcrypt.CompareHashAndPassword([]byte(u.PinHash), []byte(oldPin)); err != nil {

		// if pin is not correct, increment attempts

		if err := svc.repo.IncrementAttempts(u.ID, u.FailedPinAttempts+1); err != nil {
			return err
		}

		// if increment > 3, then freeze acc
		if u.FailedPinAttempts+1 >= 3 {

			if err := svc.repo.LockWallet(u.ID, time.Now().Add(30*time.Minute)); err != nil {
				return errors.New("Failed to lock wallet")
			}
			return errors.New("too many attempts, wallet locked for 1 minute")
		}

		return errors.New("Old Pin Not Correct")

	}

	//create new pin

	hashPin, _ := bcrypt.GenerateFromPassword([]byte(newPin), bcrypt.DefaultCost)

	hash := string(hashPin)

	if err := svc.repo.CreateWalletPin(userId, hash); err != nil {
		return err
	}

	if err := svc.repo.ResetLockout(u.ID); err != nil {
		return err
	}

	return nil

}

func (svc *WalletService) FreezeWallet(userId uuid.UUID, reason string) error {

	// get user

	userW, err := svc.repo.GetUserById(userId)

	if err != nil {
		return errors.New("Unable to get user wallet")
	}

	// Freeze
	return svc.repo.Freeze(userW.ID, reason)

}

func (svc *WalletService) StartDailyReset() {
	c := cron.New()

	c.AddFunc("0 0 * * *", func() {
		svc.repo.ResetDailySpent()
	})

	c.Start()

}

func (svc *WalletService) GetHistory(userId uuid.UUID) ([]model.Transaction, error) {

	w, err := svc.repo.GetUserById(userId)

	if err != nil {
		return nil, errors.New("Error occured getting wallet")
	}

	history, err := svc.repo.History(w.ID)

	if err != nil {
		return nil, errors.New("Error occured getting the history")
	}

	return history, nil
}

func (svc *WalletService) GetWallet(userId uuid.UUID) (model.Wallet, error) {

	w, err := svc.repo.GetUserById(userId)

	if err != nil {
		return model.Wallet{}, errors.New("Error occured getting wallet")
	}

	wal, err := svc.repo.Wallet(w.ID)

	if err != nil {
		return model.Wallet{}, errors.New("Error occured getting balance")
	}

	return wal, nil

}
