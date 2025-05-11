package controllers

import (
	"net/http"
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
	type UserSigninRequest struct {
		Email          string `form:"email" binding:"required"`
		HashedPassword string `form:"password" binding:"required"`
	}
	var req UserSigninRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	user, err := sm.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": err})
		return
	}
	if user.HashedPassword != req.HashedPassword {
		c.JSON(http.StatusNotFound, gin.H{"message": "invalid user credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user})
}

// WRITES,DELETES AND UPDATES
func UserSignup(c *gin.Context) {
	type UserSignupRequest struct {
		FirstName                string    `form:"first_name" binding:"required"`
		LastName                 string    `form:"last_name" binding:"required"`
		HashedPassword           string    `form:"password" binding:"required"`
		Email                    string    `form:"email" binding:"required"`
		DateOfBirth              time.Time `form:"dob" binding:"required"`
		IdentificationNumber     string    `form:"id_number" binding:"required"`
		IdentificationImageFront string    `form:"id_image_front" binding:"required"`
		IdentificationImageBack  string    `form:"id_image_back" binding:"required"`
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
		IdentificationImageBack: req.IdentificationImageBack, PrevPW: "", NewPW: "", UserID: -1, Action: utils.UserCreateAccount,
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
		UserID      int    `form:"user_id" binding:"required"`
		OldPassword string `form:"old_password" binding:"required"`
		NewPassword string `form:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Marshall
	payload := utils.UserPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "", DateOfBirth: time.Now(),
		IdentificationNumber: "", IdentificationImageFront: "", IdentificationImageBack: "",
		PrevPW: req.OldPassword, NewPW: req.NewPassword, UserID: req.UserID, Action: utils.UserUpdatePassword,
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
		UserID int `form:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	payload := utils.UserPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "", DateOfBirth: time.Now(),
		IdentificationNumber: "", IdentificationImageFront: "", IdentificationImageBack: "",
		PrevPW: "", NewPW: "", UserID: req.UserID, Action: utils.UserDeleteAccount,
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
		UserID int `form:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	payload := utils.UserPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "", DateOfBirth: time.Now(),
		IdentificationNumber: "", IdentificationImageFront: "", IdentificationImageBack: "",
		PrevPW: "", NewPW: "", UserID: req.UserID, Action: utils.UserDeleteAccount,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "operation pending"})
}
