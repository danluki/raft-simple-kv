package fsm

import "github.com/hashicorp/raft"

type snapshotNoop struct{}

func (s *snapshotNoop) Persist(_ raft.SnapshotSink) error { return nil }

func (s *snapshotNoop) Release() {}

func newSnapshotNoop() (raft.FSMSnapshot, error) {
	return &snapshotNoop{}, nil
}
