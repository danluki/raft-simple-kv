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

type requestStore struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Store handling save to raft cluster. Store will invoke raft.Apply to make this stored in all cluster
// with acknowledge from n quorum. Store must be done in raft leader, otherwise return error.
func (h *handler) Store(c echo.Context) error {
	var form = requestStore{}
	if err := c.Bind(&form); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error binding: %s", err.Error()),
		})
	}

	form.Key = strings.TrimSpace(form.Key)
	if form.Key == "" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is required",
		})
	}

	if h.raft.State() != raft.Leader {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": "not the leader",
		})
	}

	payload := fsm.CommandPayload{
		Operation: "SET",
		Key:       form.Key,
		Value:     form.Value,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": fmt.Sprintf("error preparing saving data payload: %s", err.Error()),
		})
	}

	applyFuture := h.raft.Apply(data, 500*time.Millisecond)
	if err := applyFuture.Error(); err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": fmt.Sprintf("error persisting data in raft cluster: %s", err.Error()),
		})
	}

	_, ok := applyFuture.Response().(*fsm.ApplyResponse)
	if !ok {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "error response is not match apply response",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success persisting data",
		"data":    form,
	})
}
