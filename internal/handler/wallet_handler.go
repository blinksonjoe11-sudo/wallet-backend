package handler

import (
	"net/http"
	"wallet-api/internal/middleware"
	"wallet-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WalletHandler struct {
	svc *service.WalletService
}

type Amounts struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type Reasons struct {
	Reason string `json:"reason" binding:"required"`
}

type UpdatePin struct {
	OldPin        string `json:"old_pin" binding:"required"`
	NewPin        string `json:"new_Pin" binding:"required"`
	ConfirmNewPin string `json:"confirm_new_Pin" binding:"required"`
}

type Transfer struct {
	Pin           string  `json:"pin" binding:"required"`
	AccountNumber string  `json:"account_number"`
	Narration     string  `json:"narration"`
	Amount        float64 `json:"amount" binding:"required"`
}

type WalletPin struct {
	Pin string `json:"pin"`
}

func NewWalletHandler(svc *service.WalletService) *WalletHandler {
	return &WalletHandler{
		svc: svc,
	}
}

func (h *WalletHandler) AddFund(c *gin.Context) {

	user, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized user",
		})
		return
	}

	var req Amounts

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	userId, err := uuid.Parse(user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user Id",
		})
		return
	}

	if err := h.svc.Fund(userId, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": "Funds added",
	})
}

func (h *WalletHandler) TransferFunds(c *gin.Context) {

	sender, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized user",
		})
		return
	}

	var transfer Transfer

	if err := c.ShouldBindJSON(&transfer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	senderId, err := uuid.Parse(sender)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid sender Id",
		})
		return
	}

	if err := h.svc.Transfer(transfer.Pin, senderId, transfer.AccountNumber, transfer.Amount, transfer.Narration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": "Transfer Successful",
	})

}

func (h *WalletHandler) CreatePin(c *gin.Context) {

	user, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized user",
		})
		return
	}

	userId, _ := uuid.Parse(user)

	var body WalletPin

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	if err := h.svc.CreateWalletPinUser(userId, body.Pin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pin Created Successfully",
	})

}

func (h *WalletHandler) UpdatePin(c *gin.Context) {

	user, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized user",
		})
		return
	}

	userId, _ := uuid.Parse(user)

	var u UpdatePin

	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Body Request",
		})
		return
	}

	if err := h.svc.UpdateWalletPinUser(userId, u.OldPin, u.NewPin, u.ConfirmNewPin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pin Updated",
	})
}

func (h *WalletHandler) FreezeUserWallet(c *gin.Context) {

	user, ok := middleware.GetUserId(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized user",
		})
		return
	}

	userId, _ := uuid.Parse(user)

	var r Reasons

	if err := c.ShouldBindJSON(&r); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := h.svc.FreezeWallet(userId, r.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account Frozen",
	})

}

func (h *WalletHandler) GetAllUserTransactions(c *gin.Context) {

	user, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized user",
		})
		return
	}

	userId, _ := uuid.Parse(user)

	data, err := h.svc.GetHistory(userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})

}

func (h *WalletHandler) GetUserWallet(c *gin.Context) {

	user, ok := middleware.GetUserId(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized user",
		})
		return
	}

	userId, _ := uuid.Parse(user)

	data, err := h.svc.GetWallet(userId)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": data,
	})

}
