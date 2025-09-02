package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func SetupRoutes(accountHandler *AccountHandler) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.POST("/accounts", accountHandler.CreateAccount)
	e.GET("/accounts", accountHandler.GetAccounts)
	e.GET("/accounts/:id", accountHandler.GetAccount)
	e.POST("/accounts/:id/deposit", accountHandler.Deposit)
	e.POST("/accounts/:id/withdraw", accountHandler.Withdraw)

	return e
}
