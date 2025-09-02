package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/viniciuslima/account-EDA/internal/application/command"
	"github.com/viniciuslima/account-EDA/internal/application/query"
)

// AccountHandler gerencia requisições HTTP relacionadas a contas
type AccountHandler struct {
	createAccountHandler *command.CreateAccountHandler
	depositHandler       *command.DepositHandler
	withdrawHandler      *command.WithdrawHandler
	accountQuery         query.AccountQuery
}

// NewAccountHandler cria um novo manipulador de contas
func NewAccountHandler(
	createAccountHandler *command.CreateAccountHandler,
	depositHandler *command.DepositHandler,
	withdrawHandler *command.WithdrawHandler,
	accountQuery query.AccountQuery,
) *AccountHandler {
	return &AccountHandler{
		createAccountHandler: createAccountHandler,
		depositHandler:       depositHandler,
		withdrawHandler:      withdrawHandler,
		accountQuery:         accountQuery,
	}
}

// CreateAccount manipula requisições para criar uma conta
func (h *AccountHandler) CreateAccount(c echo.Context) error {
	var cmd command.CreateAccountCommand
	if err := c.Bind(&cmd); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	accountID, err := h.createAccountHandler.Handle(cmd)
	if err != nil {
		if err == command.ErrEmailAlreadyExists {
			return c.JSON(http.StatusConflict, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": accountID})
}

// GetAccount manipula requisições para buscar uma conta pelo ID
func (h *AccountHandler) GetAccount(c echo.Context) error {
	id := c.Param("id")

	account, err := h.accountQuery.GetByID(id)
	if err != nil {
		if err == query.ErrAccountNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, account)
}

// GetAccounts manipula requisições para listar todas as contas
func (h *AccountHandler) GetAccounts(c echo.Context) error {
	accounts, err := h.accountQuery.GetAll()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, accounts)
}

// Deposit manipula requisições para depositar em uma conta
func (h *AccountHandler) Deposit(c echo.Context) error {
	id := c.Param("id")

	var cmd command.DepositCommand
	if err := c.Bind(&cmd); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	cmd.AccountID = id

	if err := h.depositHandler.Handle(cmd); err != nil {
		if err == command.ErrAccountNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if err == command.ErrInvalidAmount {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Busca a conta atualizada
	account, err := h.accountQuery.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, account)
}

// Withdraw manipula requisições para sacar de uma conta
func (h *AccountHandler) Withdraw(c echo.Context) error {
	id := c.Param("id")

	var cmd command.WithdrawCommand
	if err := c.Bind(&cmd); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	cmd.AccountID = id

	if err := h.withdrawHandler.Handle(cmd); err != nil {
		if err == command.ErrAccountNotFound {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Account not found"})
		}
		if err == command.ErrInvalidAmount || err == command.ErrInsufficientFunds {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Busca a conta atualizada
	account, err := h.accountQuery.GetByID(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, account)
}
