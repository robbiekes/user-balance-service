package v1

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"user-balance-api/internal/entity"
	"user-balance-api/internal/service"
)

const (
	ALL_HISTORY_DATA = "all history"
)

type historyRoutes struct {
	s service.History
}

func newHistoryRoutes(g *echo.Group, s service.History) {
	h := &historyRoutes{s: s}

	g.GET("/all", h.getAll)  // + ?sort={category}; + ?limit=5&cursor=
	g.GET("/:id", h.getById) // + ?sort={category}; + ?limit=5&cursor=

}

type PaginationArgs struct {
	Limit     int
	Param     string
	AccountId int
}

type SortArgs struct {
	Type      string
	AccountId int
}

func (h *historyRoutes) getAll(c echo.Context) error {
	sort := c.FormValue("sort")

	limit, _ := strconv.Atoi(c.FormValue("limit"))
	pgn := PaginationArgs{
		Limit:     limit,
		Param:     c.FormValue("cursor"),
		AccountId: 0,
	}

	srt := SortArgs{
		Type:      sort,
		AccountId: 0,
	}

	var records []entity.History
	var err error

	if limit > 0 {
		records, err = h.s.Pagination(c.Request().Context(), pgn)
	} else if len(sort) != 0 && (sort == "date" || sort == "amount") {
		records, err = h.s.ShowSorted(c.Request().Context(), srt)
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
	accountId, _ := strconv.Atoi(c.Param("id"))
	limit, _ := strconv.Atoi(c.FormValue("limit"))
	pgn := PaginationArgs{
		Limit:     limit,
		Param:     c.FormValue("cursor"),
		AccountId: accountId,
	}

	srt := SortArgs{
		Type:      c.FormValue("sort"),
		AccountId: accountId,
	}

	var records []entity.History
	var err error

	if limit > 0 {
		records, err = h.s.Pagination(c.Request().Context(), pgn)
	} else if len(srt.Type) != 0 && (srt.Type == "date" || srt.Type == "amount") {
		records, err = h.s.ShowSorted(c.Request().Context(), srt)
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
