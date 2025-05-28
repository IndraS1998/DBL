package controllers

import (
	"net/http"
	"raft/state"
	sm "raft/state/stateMachine"
	"raft/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// READS
func Pong(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Pong"})
}

func GetLogEntry(c *gin.Context) {
	poll_id := c.Query("poll_id")
	logEntry, err := state.GetLogEntryForApi(poll_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Entry": logEntry})
}

func GetWalletsCount(c *gin.Context) {
	id := c.Query("id")
	user_id, err := strconv.Atoi(id)
	if err == nil {
		wallets, err := sm.CountWalletsByUser(user_id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{"wallets": wallets})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user ID"})
		return
	}
}

func GetGlobalBalance(c *gin.Context) {
	id := c.Query("id")
	user_id, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"message": "invalid user ID"})
		return
	}
	balance, err := sm.SumWalletBallancesByUser(user_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": balance})
}

func GetTransactions(c *gin.Context) {
	id := c.Query("id")
	month := c.Query("month")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	start, end, err := ParseMonth(month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format"})
		return
	}
	count, err := sm.CountUserTransactionBetween(userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": count})
}

func GetTransactionsSum(c *gin.Context) {
	id := c.Query("id")
	month := c.Query("month")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}
	start, end, err := ParseMonth(month)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month format"})
		return
	}
	sum, err := sm.SumUserTransactionBetween(userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sum": sum})
}

func GetUserInfo(c *gin.Context) {
	uid := c.Query("user_id")
	userID, err := strconv.Atoi(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid user ID"})
		return
	}
	user, err := sm.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UserSignin(c *gin.Context) {
	email := c.Query("email")
	password := c.Query("password")

	user, err := sm.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err})
		return
	}
	if user.HashedPassword != password {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid user credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// WRITES,DELETES AND UPDATES
func UserSignup(c *gin.Context) {
	type UserSignupRequest struct {
		FirstName                string    `json:"first_name" binding:"required"`
		LastName                 string    `json:"last_name" binding:"required"`
		HashedPassword           string    `json:"password" binding:"required"`
		Email                    string    `json:"email" binding:"required"`
		DateOfBirth              time.Time `json:"dob" binding:"required"`
		IdentificationNumber     string    `json:"id_number" binding:"required"`
		IdentificationImageFront string    `json:"id_image_front" binding:"required"`
		IdentificationImageBack  string    `json:"id_image_back" binding:"required"`
		PollID                   string    `json:"poll_id" binding:"required"`
	}
	var req UserSignupRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	// Marshall
	payload := utils.UserPayload{
		FirstName: req.FirstName, LastName: req.LastName, HashedPassword: req.HashedPassword, Email: req.Email,
		DateOfBirth:          req.DateOfBirth,
		IdentificationNumber: req.IdentificationNumber, IdentificationImageFront: req.IdentificationImageFront,
		IdentificationImageBack: req.IdentificationImageBack, PrevPW: "", NewPW: "",
		UserID: -1, Action: utils.UserCreateAccount, PollID: req.PollID,
	}
	//add payload to redis queue
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	//TODO :  we need to follow raft protocol here (idea generate a uid that will be used eventually to get status updates on the operation)
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}

func UpdatePassword(c *gin.Context) {
	var req struct {
		UserID      int    `json:"user_id" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
		PollID      string `json:"poll_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Marshall
	payload := utils.UserPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "", DateOfBirth: time.Now(),
		IdentificationNumber: "", IdentificationImageFront: "", IdentificationImageBack: "",
		PrevPW: req.OldPassword, NewPW: req.NewPassword, UserID: req.UserID, Action: utils.UserUpdatePassword, PollID: req.PollID,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}

func DeleteUser(c *gin.Context) {
	var req struct {
		UserID int    `json:"user_id" binding:"required"`
		PollID string `json:"poll_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	payload := utils.UserPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "", DateOfBirth: time.Now(),
		IdentificationNumber: "", IdentificationImageFront: "", IdentificationImageBack: "",
		PrevPW: "", NewPW: "", UserID: req.UserID, Action: utils.UserDeleteAccount, PollID: req.PollID,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}

func CreateWallet(c *gin.Context) {
	var req struct {
		UserID int    `json:"user_id" binding:"required"`
		PollID string `json:"poll_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	payload := utils.UserPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "", DateOfBirth: time.Now(),
		IdentificationNumber: "", IdentificationImageFront: "", IdentificationImageBack: "",
		PrevPW: "", NewPW: "", UserID: req.UserID, Action: utils.UserCreateWallet, PollID: req.PollID,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}
