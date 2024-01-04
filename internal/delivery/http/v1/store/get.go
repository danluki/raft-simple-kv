package store

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *handler) Get(c echo.Context) error {
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		return c.JSON(http.StatusUnprocessableEntity, map[string]interface{}{
			"error": "key is required",
		})
	}

	var keyByte = []byte(key)

	txn := h.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get(keyByte)
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": err.Error(),
		})
	}

	value := []byte{}
	err = item.Value(func(val []byte) error {
		value = append(value, val...)
		return nil
	})

	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": err.Error(),
		})
	}

	var data interface{}
	if len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}
	if err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success fetching data",
		"data": map[string]interface{}{
			"key":   key,
			"value": data,
		},
	})
}
