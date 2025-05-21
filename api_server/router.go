package api_server

import (
	"raft/api_server/controllers"
	"raft/state"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, node *state.Node) {

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // or "*" in development
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(LeaderOnly(node))
	r.GET("/ping", controllers.Pong)
	r.GET("/log", controllers.GetLogEntry)
	user := r.Group("/api/user")
	{
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
		admin.GET("/signin", controllers.AdminSignin)
		admin.POST("/signup", controllers.AdminSignup)
		admin.POST("/validate/user", controllers.ValidateUser)
	}

	stats := r.Group("/api/admin/stats")
	{
		stats.GET("/active-users", controllers.CountActiveUsers)
		stats.GET("/count/transactions/", controllers.CountTransactionsForMonth)
		stats.GET("/sum/transactions/", controllers.SumTransactionsForMonth)
		stats.GET("/wallets/count", controllers.CountWallets)
		stats.GET("/transactions/recent", controllers.GetRecentTransactions)
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
