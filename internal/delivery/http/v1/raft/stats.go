package raft

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *handler) Stats(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Here is the raft status",
		"data":    h.raft.Stats(),
	})
}
