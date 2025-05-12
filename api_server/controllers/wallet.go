package controllers

import (
	"net/http"
	sm "raft/state/stateMachine"
	"raft/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// READS
func GetWalletInfo(c *gin.Context) {
	wid := c.Query("wallet_id")
	walletID, err := strconv.Atoi(wid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid wallet ID"})
		return
	}
	wallet, err := sm.GetWallet(walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallet": wallet})
}

func GetWalletsByUser(c *gin.Context) {
	uid := c.Query("user_id")
	userID, err := strconv.Atoi(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	wallets, err := sm.GetWallets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallets": wallets})
}

// MODIFICATIONS
func Transfer(c *gin.Context) {
	type transferData struct {
		Sender_wallet_id   int   `json:"sender_wallet_id" binding:"required"`
		Receiver_wallet_id int   `json:"receiver_wallet_id" binding:"required"`
		Amount             int64 `json:"amount" binding:"required"`
	}

	var req transferData

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := utils.WalletOperationPayload{
		Wallet1: req.Sender_wallet_id,
		Wallet2: req.Receiver_wallet_id,
		Amount:  req.Amount,
		Action:  utils.WalletTransfer,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}

func Withdraw(c *gin.Context) {
	type withdrawData struct {
		Wallet_id int   `json:"sender_wallet_id" binding:"required"`
		Amount    int64 `json:"amount" binding:"required"`
	}
	var req withdrawData
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := utils.WalletOperationPayload{
		Wallet1: req.Wallet_id,
		Wallet2: -1,
		Amount:  req.Amount,
		Action:  utils.WalletWithdraw,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}

func Deposit(c *gin.Context) {
	type depositData struct {
		Wallet_id int   `json:"sender_wallet_id" binding:"required"`
		Amount    int64 `json:"amount" binding:"required"`
	}
	var req depositData
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload := utils.WalletOperationPayload{
		Wallet1: req.Wallet_id,
		Wallet2: -1,
		Amount:  req.Amount,
		Action:  utils.WalletDeposit,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}
