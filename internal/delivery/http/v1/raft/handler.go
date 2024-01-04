package raft

import (
	"github.com/hashicorp/raft"
)

type handler struct {
	raft *raft.Raft
}

func NewHandler(raft *raft.Raft) *handler {
	return &handler{
		raft: raft,
	}
}
