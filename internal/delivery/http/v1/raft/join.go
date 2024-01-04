package raft

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
)

type requestJoin struct {
	NodeID      string `json:"node_id"`
	RaftAddress string `json:"raft_address"`
}

func (h *handler) Join(c echo.Context) error {
	form := requestJoin{}
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

	f := h.raft.AddVoter(raft.ServerID(form.NodeID), raft.ServerAddress(form.RaftAddress), 0, 0)
	if f.Error() != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": f.Error().Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("node %s at %s joined successfully", form.NodeID, form.RaftAddress),
		"data":    h.raft.Stats(),
	})
}
