package v1

import (
	raftv1 "github.com/danluki/kv-raft/internal/delivery/http/v1/raft"
	storev1 "github.com/danluki/kv-raft/internal/delivery/http/v1/store"
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
	"github.com/labstack/echo/v4"
)

type handler struct {
	raft *raft.Raft
	db   *badger.DB
}

type Opts struct {
	Raft *raft.Raft
	DB   *badger.DB
}

func NewHandler(opts Opts) *handler {
	return &handler{
		raft: opts.Raft,
		db:   opts.DB,
	}
}

func (h *handler) Init(api *echo.Group) {
	v1 := api.Group("/v1")
	{
		raftHandler := raftv1.NewHandler(h.raft)
		v1.POST("/raft/join", raftHandler.Join)
		v1.POST("/raft/remove", raftHandler.Remove)
		v1.GET("/raft/stats", raftHandler.Stats)

		storeHandler := storev1.NewHandler(h.raft, h.db)
		v1.POST("/store", storeHandler.Store)
		v1.DELETE("/store/:key", storeHandler.Delete)
		v1.GET("/store/:key", storeHandler.Get)
	}
}
