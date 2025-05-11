package api_server

import (
	"raft/api_server/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	user := r.Group("/api/user")
	{
		user.GET("/ping", controllers.Pong)
		user.GET("/", controllers.GetUserInfo)
		user.POST("/signup", controllers.UserSignup)
		user.POST("/login", controllers.UserSignin)
		user.PATCH("/", controllers.UpdatePassword)
		user.DELETE("/", controllers.DeleteUser)
	}

	admin := r.Group("/api/admin")
	{
		admin.GET("/", controllers.GetAdminInfo)
		admin.GET("/users", controllers.GetAllUsers)
		admin.POST("/signup", controllers.AdminSignup)
		admin.POST("/signin", controllers.AdminSignin)
		admin.POST("/validate/user", controllers.ValidateUser)
	}

	wallet := r.Group("/api/wallet")
	{
		wallet.GET("/", controllers.GetWalletInfo)
		wallet.GET("/user", controllers.GetWalletsByUser)
		wallet.POST("/create", controllers.CreateWallet)
		wallet.POST("/transfer", controllers.Transfer)
		wallet.POST("/deposit", controllers.Deposit)
		wallet.POST("/withdraw", controllers.Withdraw)
	}

}
