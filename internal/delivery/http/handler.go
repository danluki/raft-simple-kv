package http

import (
	"net/http"
	"time"

	v1 "github.com/danluki/kv-raft/internal/delivery/http/v1"
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Handler struct {
	raft *raft.Raft
	db   *badger.DB
}

type Opts struct {
	Raft *raft.Raft
	DB   *badger.DB
}

func NewHandler(opts Opts) *Handler {
	return &Handler{
		raft: opts.Raft,
		db:   opts.DB,
	}
}

func (h *Handler) Init() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	// e.HidePort = true
	e.Logger.SetLevel(log.INFO)
	e.Pre(middleware.RemoveTrailingSlash())
	e.GET("/", func(c echo.Context) error {
		time.Sleep(5 * time.Second)
		return c.JSON(http.StatusOK, "OK")
	})
	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))

	h.initApi(e)

	return e
}

func (h *Handler) initApi(e *echo.Echo) {
	handlerV1 := v1.NewHandler(v1.Opts{
		Raft: h.raft,
		DB:   h.db,
	})
	api := e.Group("/api")
	{
		handlerV1.Init(api)
	}
}
