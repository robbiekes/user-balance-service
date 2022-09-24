package v1

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
	"user-balance-service/internal/entity"
	"user-balance-service/internal/service"
)

const (
	refillType        = "пополнение счёта"
	writeoffType      = "снятие со счёта"
	innertransferType = "иcходящий перевод"
	outertransferType = "входящий перевод"
)

type accountRoutes struct {
	s service.Account
	h service.History
}

func newAccountRoutes(g *echo.Group, s service.Account, h service.History) {
	r := &accountRoutes{s, h}

	g.POST("/create", r.createAccount)
	g.GET("/state", r.getBalance) // ?currency=USD to get balance in chosen currency
	g.PUT("/refill", r.refillBalance)
	g.PUT("/write-off", r.writeOffBalance)
	g.PUT("/transfer", r.transferMoney)
	g.DELETE("/delete", r.deleteAccount)
}

// create account and set balance to 0
func (r *accountRoutes) createAccount(c echo.Context) error {
	var input entity.Account

	err := c.Bind(&input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	id, err := r.s.CreateAccount(c.Request().Context())
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

// refill balance
func (r *accountRoutes) refillBalance(c echo.Context) error {
	var input entity.Account

	err := c.Bind(&input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	err = r.s.MakeDeposit(c.Request().Context(), input.Id, input.Balance)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	_, err = r.h.SaveHistory(c.Request().Context(), entity.History{
		Type:        refillType,
		Description: "",
		Amount:      input.Balance,
		AccountId:   input.Id,
		Date:        entity.CustomTime(time.Now()),
	})
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

//  write off some money
func (r *accountRoutes) writeOffBalance(c echo.Context) error {
	var input entity.Account

	err := c.Bind(&input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	err = r.s.WriteOff(c.Request().Context(), input.Id, input.Balance)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	_, err = r.h.SaveHistory(c.Request().Context(), entity.History{
		Type:        writeoffType,
		Description: "",
		Amount:      input.Balance,
		AccountId:   input.Id,
		Date:        entity.CustomTime(time.Now()),
	})
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

type TransferRequest struct {
	IdFrom int `json:"id_from"`
	IdTo   int `json:"id_to"`
	Amount int `json:"amount"`
}

// transfer money from one account to another
func (r *accountRoutes) transferMoney(c echo.Context) error {
	var transaction TransferRequest

	err := c.Bind(&transaction)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	err = r.s.TransferMoney(c.Request().Context(), transaction.IdFrom, transaction.IdTo, transaction.Amount)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	_, err = r.h.SaveHistory(c.Request().Context(), entity.History{
		Type:        innertransferType,
		Description: "",
		Amount:      transaction.Amount,
		AccountId:   transaction.IdFrom,
		Date:        entity.CustomTime(time.Now()),
	})
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	_, err = r.h.SaveHistory(c.Request().Context(), entity.History{
		Type:        outertransferType,
		Description: "",
		Amount:      transaction.Amount,
		AccountId:   transaction.IdTo,
		Date:        entity.CustomTime(time.Now()),
	})
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

// check balance in rubles and convert to chosen currency
func (r *accountRoutes) getBalance(c echo.Context) error {
	currency := c.FormValue("currency")
	var input entity.Account
	var balance float64

	err := c.Bind(&input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	output, err := r.s.GetAccount(c.Request().Context(), input.Id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	if len(currency) != 0 {
		item, err := r.s.ConvertToCurrency(c.Request().Context(), currency, float64(output.Balance))
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return err
		}
		balance = float64(output.Balance) / item
	} else {
		balance = float64(output.Balance)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return err
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"balance": balance,
	})
}

func (r *accountRoutes) deleteAccount(c echo.Context) error {
	var input entity.Account

	err := c.Bind(&input)
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return err
	}

	err = r.s.DeleteAccount(c.Request().Context(), input.Id)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
