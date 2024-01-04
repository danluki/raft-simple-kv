package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

type badgerFSM struct {
	db *badger.DB
}

func NewBadger(badgerDB *badger.DB) raft.FSM {
	return &badgerFSM{db: badgerDB}
}

func (b *badgerFSM) get(key string) (interface{}, error) {
	var keyByte = []byte(key)
	var data interface{}

	txn := b.db.NewTransaction(false)
	defer func() {
		_ = txn.Commit()
	}()

	item, err := txn.Get(keyByte)
	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	value := make([]byte, 0)
	err = item.Value(func(v []byte) error {
		value = append(value, v...)
		return nil
	})

	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	if len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}

	if err != nil {
		data = map[string]interface{}{}
	}

	return data, err
}

func (b *badgerFSM) set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if data == nil || len(data) <= 0 {
		return nil
	}

	txn := b.db.NewTransaction(true)
	err = txn.Set([]byte(key), data)
	if err != nil {
		txn.Discard()
		return err
	}

	return txn.Commit()
}

func (b *badgerFSM) delete(key string) error {
	var keyByte = []byte(key)

	txn := b.db.NewTransaction(true)
	defer func() {
		_ = txn.Commit()
	}()

	err := txn.Delete(keyByte)
	if err != nil {
		return err
	}

	return nil
}

// Apply log is invoked once a log entry is commited
// It returns a value which will be made available in the
// ApplyFuture returned by Rafy.Apply method if that nethod called on the same Raft node as the FSM.
func (b *badgerFSM) Apply(log *raft.Log) interface{} {
	switch log.Type {
	case raft.LogCommand:
		var payload = CommandPayload{}
		if err := json.Unmarshal(log.Data, &payload); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error marshalling store payload %s\n", err.Error())
			return nil
		}

		op := strings.ToUpper(strings.TrimSpace(payload.Operation))
		switch op {
		case "SET":
			return &ApplyResponse{
				Error: b.set(payload.Key, payload.Value),
				Data:  payload.Value,
			}
		case "GET":
			data, err := b.get(payload.Key)
			return &ApplyResponse{
				Error: err,
				Data:  data,
			}

		case "DELETE":
			return &ApplyResponse{
				Error: b.delete(payload.Key),
				Data:  payload.Value,
			}
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "not raft log command type\n")
	return nil
}

// Snapshot will be called during make snapshot.
// Snapshot is used to support log compaction.
// No need to call snapshot since it already persisted in disk (using BadgerDB) when raft calling Apply function.
func (b *badgerFSM) Snapshot() (raft.FSMSnapshot, error) {
	return newSnapshotNoop()
}

// Restore is used to restore an FSM from a Snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
// Restore will update all data in BadgerDB
func (b *badgerFSM) Restore(rClose io.ReadCloser) error {
	defer func() {
		if err := rClose.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[FINALLY RESTORE] close error %s\n", err.Error())
		}
	}()

	_, _ = fmt.Fprintf(os.Stdout, "[START RESTORE] read all message from snapshot\n")
	var totalRestored int

	decoder := json.NewDecoder(rClose)
	for decoder.More() {
		var data = &CommandPayload{}
		err := decoder.Decode(data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] decode error %s\n", err.Error())
		}

		if err := b.set(data.Key, data.Value); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] set error %s\n", err.Error())
		}

		totalRestored++
	}

	// read closing bracket
	_, err := decoder.Token()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error %s\n", err.Error())
		return err
	}

	_, _ = fmt.Fprintf(
		os.Stdout,
		"[END RESTORE] success restore %d messages in snapshot\n",
		totalRestored,
	)
	return nil
}
