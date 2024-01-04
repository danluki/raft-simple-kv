package raft

import "github.com/labstack/echo/v4"

type requestRemove struct {
	NodeID string `json:"node_id"`
}

func (h *handler) Remove(c echo.Context) error {
	form := requestRemove{}
}
