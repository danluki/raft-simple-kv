package raft

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
)

type requestRemove struct {
	NodeID string `json:"node_id"`
}

func (h *handler) Remove(c echo.Context) error {
	form := requestRemove{}
	if err := c.Bind(&form); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	if h.raft.State() != raft.Leader {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": "not the leader",
		})
	}

	configFuture := h.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": err.Error(),
		})
	}

	f := h.raft.RemoveServer(raft.ServerID(form.NodeID), 0, 0)
	if err := f.Error(); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("node %s removed successfully", form.NodeID),
		"data":    h.raft.Stats(),
	})
}
