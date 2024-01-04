package store

import (
	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

type handler struct {
	raft *raft.Raft
	db   *badger.DB
}

func NewHandler(raft *raft.Raft, db *badger.DB) *handler {
	return &handler{
		raft: raft,
		db:   db,
	}
}
