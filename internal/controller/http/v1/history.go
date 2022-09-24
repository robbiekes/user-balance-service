package v1

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"user-balance-service/internal/entity"
	"user-balance-service/internal/service"
)

type historyRoutes struct {
	s service.History
}

func newHistoryRoutes(g *echo.Group, s service.History) {
	h := &historyRoutes{s: s}

	g.GET("/all", h.getAll)  // + ?sort={category}; + ?limit=5&cursor=
	g.GET("/:id", h.getById) // + ?sort={category}; + ?limit=5&cursor=

}

func (h *historyRoutes) getAll(c echo.Context) error {
	var records []entity.History
	var limit int
	var err error

	sort := c.FormValue("sort")
	param := c.FormValue("cursor")
	limitCheckNum := c.FormValue("limit")
	if len(limitCheckNum) == 0 {
		limit = 0
	} else {
		limit, err = strconv.Atoi(c.FormValue("limit"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "v1 - history - getAll - strconv.Atoi(c.FormValue())")
			return err
		}
	}

	if limit > 0 {
		records, err = h.s.Pagination(c.Request().Context(), limit, param, 0)
	} else if len(sort) != 0 && (sort == "date" || sort == "amount") {
		records, err = h.s.ShowSorted(c.Request().Context(), sort, 0)
	} else {
		records, err = h.s.ShowAll(c.Request().Context())
	}
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"records": records,
	})
}

func (h *historyRoutes) getById(c echo.Context) error {
	var records []entity.History
	var limit int
	var err error

	sort := c.FormValue("sort")
	param := c.FormValue("cursor")
	accountId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "v1 - history - getById - strconv.Atoi(c.Param())")
		return err
	}
	limitCheckNum := c.FormValue("limit")
	if len(limitCheckNum) == 0 {
		limit = 0
	} else {
		limit, err = strconv.Atoi(c.FormValue("limit"))
		if err != nil {
			newErrorResponse(c, http.StatusBadRequest, "v1 - history - getById - strconv.Atoi(c.FormValue())")
			return err
		}
	}

	if limit > 0 {
		records, err = h.s.Pagination(c.Request().Context(), limit, param, accountId)
	} else if len(sort) != 0 && (sort == "date" || sort == "amount") {
		records, err = h.s.ShowSorted(c.Request().Context(), sort, accountId)
	} else {
		records, err = h.s.ShowById(c.Request().Context(), accountId)
	}

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return err
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"records": records,
	})
}
