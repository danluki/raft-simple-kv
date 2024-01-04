package store

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danluki/kv-raft/internal/fsm"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
)

func (h *handler) Delete(c echo.Context) error {
	var key = strings.TrimSpace(c.Param("key"))
	if key == "" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is empty",
		})
	}

	if h.raft.State() != raft.Leader {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "not the leader",
		})
	}

	payload := fsm.CommandPayload{
		Operation: "DELETE",
		Key:       key,
		Value:     nil,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error preparing remove data payload: %s", err.Error()),
		})
	}

	applyFuture := h.raft.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error removing data in raft cluster: %s", err.Error()),
		})
	}

	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error response is not match apply response"),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success removing data",
		"data": map[string]interface{}{
			"key":   key,
			"value": nil,
		},
	})
}
