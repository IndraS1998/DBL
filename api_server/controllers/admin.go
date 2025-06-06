package controllers

import (
	"fmt"
	"net/http"
	"raft/state"
	sm "raft/state/stateMachine"
	"raft/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// READS
func GetAdminInfo(c *gin.Context) {
	aid := c.Query("admin_id")
	adminID, err := strconv.Atoi(aid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid admin ID"})
		return
	}
	admin, err := sm.GetAdminInfo(adminID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid admin ID"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"admin": admin})
}

func GetAllUsers(c *gin.Context) {
	users, err := sm.GetUsers()
	if err != nil {
		fmt.Println("fetch all users error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"Error": err})
	} else {
		c.JSON(http.StatusOK, gin.H{"Users": users})
	}

}

func AdminSignin(c *gin.Context) {
	email := c.Query("email")
	password := c.Query("password")

	admin, err := sm.GetAdminByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": err})
		return
	}
	if admin.HashedPassword != password {
		c.JSON(http.StatusNotAcceptable, gin.H{"message": "invalid credentials"})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"admin": admin})
}

// CountActiveUsers returns the total number of validated (active) users
func CountActiveUsers(c *gin.Context) {
	count, err := sm.CountValidatedUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count active users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"active_users": count})
}

// CountTransactionsForMonth returns the number of transactions for a given month (YYYY-MM)
func CountTransactionsForMonth(c *gin.Context) {
	month := c.Query("month") // expected format: YYYY-MM

	start, end, err := ParseMonth(month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format"})
		return
	}

	count, err := sm.CountWalletOperationsBetween(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count transactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction_count": count})
}

// SumTransactionsForMonth returns the total sum of transaction amounts for a given month (YYYY-MM)
func SumTransactionsForMonth(c *gin.Context) {
	month := c.Query("month") // expected format: YYYY-MM
	start, end, err := ParseMonth(month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format"})
		return
	}

	sum, err := sm.SumWalletOperationAmountsBetween(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sum transactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total_amount": sum})
}

// CountWallets returns the total number of wallets
func CountWallets(c *gin.Context) {
	count, err := sm.CountWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count wallets"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wallet_count": count})
}

// GetRecentTransactions returns the 5 most recent wallet operations
func GetRecentTransactions(c *gin.Context) {
	operations, err := sm.GetMostRecentWalletOperations(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch recent transactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recent_transactions": operations})
}

// parseMonth parses "YYYY-MM" into start and end time.Time objects
func ParseMonth(month string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

// MODIFICATIONS
func AdminSignup(c *gin.Context) {
	type AdminSignupPayload struct {
		FirstName      string `json:"first_name" binding:"required"`
		LastName       string `json:"last_name" binding:"required"`
		HashedPassword string `json:"hashed_password" binding:"required"`
		Email          string `json:"email" binding:"required"`
		PollID         string `json:"poll_id" binding:"required"`
	}
	var req AdminSignupPayload
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	ct, err := state.GetCurrentTermFromAPI()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	payload := utils.AdminPayload{
		FirstName: req.FirstName, LastName: req.LastName, HashedPassword: req.HashedPassword, Email: req.Email,
		AdminID: -1, UserId: -1, Action: utils.AdminCreateAccount, PollID: req.PollID, Term: ct,
	}
	err = utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"message": "operation pending"})
}

func ValidateUser(c *gin.Context) {
	type AdminValidationPayload struct {
		AdminId int    `json:"admin_id" binding:"required"`
		UserID  int    `json:"user_id" binding:"required"`
		PollID  string `json:"poll_id" binding:"required"`
	}
	var req AdminValidationPayload
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	ct, err := state.GetCurrentTermFromAPI()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	payload := utils.AdminPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "", Term: ct,
		AdminID: req.AdminId, UserId: req.UserID, Action: utils.AdminValidateUser, PollID: req.PollID,
	}
	err = utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}
