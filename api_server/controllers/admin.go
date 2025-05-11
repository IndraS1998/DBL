package controllers

import (
	sm "raft/state/stateMachine"
	"raft/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// READS
func GetAdminInfo(c *gin.Context) {
	aid := c.Query("admin_id")
	adminID, err := strconv.Atoi(aid)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid admin ID"})
		return
	}
	admin, err := sm.GetAdminInfo(adminID)
	if err != nil {
		c.JSON(400, gin.H{"message": "invalid admin ID"})
		return
	}
	c.JSON(200, gin.H{"admin": admin})
}

func GetAllUsers(c *gin.Context) {
	users, err := sm.GetUsers()
	if err != nil {
		c.JSON(400, gin.H{"Error": err})
	} else {
		c.JSON(200, gin.H{"Users": users})
	}

}

func AdminSignin(c *gin.Context) {
	type AdminSigninPayload struct {
		HashedPassword string `form:"hashed_password" binding:"required"`
		Email          string `form:"email" binding:"required"`
	}
	var req AdminSigninPayload
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"message": err})
		return
	}
	admin, err := sm.GetAdminByEmail(req.Email)
	if err != nil {
		c.JSON(401, gin.H{"message": err})
		return
	}
	if admin.HashedPassword != req.HashedPassword {
		c.JSON(401, gin.H{"message": "invalid credentials"})
		return
	}
	c.JSON(200, gin.H{"admin": admin})
}

// MODIFICATIONS
func AdminSignup(c *gin.Context) {
	type AdminSignupPayload struct {
		FirstName      string `form:"first_name" binding:"required"`
		LastName       string `form:"last_name" binding:"required"`
		HashedPassword string `form:"hashed_password" binding:"required"`
		Email          string `form:"email" binding:"required"`
	}
	var req AdminSignupPayload
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	payload := utils.AdminPayload{
		FirstName: req.FirstName, LastName: req.LastName, HashedPassword: req.HashedPassword, Email: req.Email,
		AdminID: -1, UserId: -1, Action: utils.AdminCreateAccount,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"message": "operation pending"})
}

func ValidateUser(c *gin.Context) {
	type AdminValidationPayload struct {
		AdminId int `form:"admin_id" binding:"required"`
		UserID  int `form:"user_id" binding:"required"`
	}
	var req AdminValidationPayload
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	payload := utils.AdminPayload{
		FirstName: "", LastName: "", HashedPassword: "", Email: "",
		AdminID: req.AdminId, UserId: req.UserID, Action: utils.AdminValidateUser,
	}
	err := utils.AppendRedisPayload(payload)
	if err != nil {
		c.JSON(400, gin.H{"error": err})
		return
	}
	c.JSON(200, gin.H{"message": "operation pending"})
}
