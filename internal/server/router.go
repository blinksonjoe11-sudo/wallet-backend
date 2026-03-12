package server

import (
	"net/http"
	"wallet-api/internal/app"
	"wallet-api/internal/config"
	"wallet-api/internal/handler"
	"wallet-api/internal/middleware"
	"wallet-api/internal/repo"
	"wallet-api/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(a *app.App, cfg config.Config) *gin.Engine {

	r := gin.New()

	r.Use(gin.Logger())

	r.Use(gin.Recovery())

	r.GET("/health", health)

	userRepo := repo.NewUserRepo(a)
	walletRepo := repo.NewWalletRepo(a)

	userService := service.NewUserService(userRepo, walletRepo, a.Config.JWTSecret)
	userHandler := handler.NewUserHandler(userService)

	walletService := service.NewWalletService(walletRepo)
	walletService.StartDailyReset()
	walletHandler := handler.NewWalletHandler(walletService)

	r.POST("/register", userHandler.RegisterNewUser)
	r.POST("/login", userHandler.LoginUser)

	authenticated := r.Group("")

	authenticated.Use(middleware.AuthRequired(cfg.JWTSecret))

	authenticated.POST("/create/pin", walletHandler.CreatePin)
	authenticated.POST("/fund", walletHandler.AddFund)
	authenticated.POST("/transfer", walletHandler.TransferFunds)
	authenticated.POST("/freeze", walletHandler.FreezeUserWallet)
	authenticated.PUT("/update/pin", walletHandler.UpdatePin)
	authenticated.GET("/history", walletHandler.GetAllUserTransactions)
	authenticated.GET("/get/wallet", walletHandler.GetUserWallet)

	authenticated.GET("/wallet", func(c *gin.Context) {

		userId, _ := middleware.GetUserId(c)
		c.JSON(http.StatusOK, gin.H{
			"status": true,
			"userId": userId,
		})
	})

	return r
}
