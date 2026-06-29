package luks2

import (
	"context"
	"fmt"
	"github.com/snapcore/secboot"
	luks2 "github.com/snapcore/secboot/internal/luks2"
	"github.com/snapcore/secboot/internal/luksview"
)

// reencryptionImpl is an implementation of [secboot.Reencryption].
type reencryptionImpl struct {
	path string
}

func (r reencryptionImpl) Status() (*secboot.ReencryptionStatus, error) {
	view, err := luksview.NewView(context.TODO(), r.path)
	if err != nil {
		return nil, fmt.Errorf("cannot obtain LUKS header view: %w", err)
	}
	nReencrypt := view.HasReencrypt()
	if nReencrypt > 0 {
		var status secboot.ReencryptionStatus = secboot.ReencryptionStatusInProgress
		return &status, nil
		// TODO distinguish in-progress / interrupted?
		//return secboot.ReencryptionStatusInterrupted, nil
	} else {
		var status secboot.ReencryptionStatus = secboot.ReencryptionStatusNone
		return &status, nil
	}
}

func (r reencryptionImpl) Initialize(unlockKeys map[string][]byte) error {
	var unlockKeysOrdered [][]byte
	// TODO: keys must be sorted by their keyslot index
	for _, unlockKey := range unlockKeys {
		unlockKeysOrdered = append(unlockKeysOrdered, unlockKey)
	}
	return luks2.ReencryptInitialize(r.path, unlockKeysOrdered)
}

func (r reencryptionImpl) Resume(unlockKey []byte) (<-chan secboot.ReencryptionProgressEvent, error) {
	outChan := make(chan secboot.ReencryptionProgressEvent)
	go func() {
		defer close(outChan)
		outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressStarted}
		outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressCompleted}
	}()
	return outChan, nil
}
