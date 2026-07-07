package secboot

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrReencryptionNotActive = errors.New("not active")
)

type ReencryptionStatus int

const (
	ReencryptionStatusNone ReencryptionStatus = iota
	ReencryptionStatusInitialized
)

type ReencryptionProgressEventType int

const (
	ReencryptionProgressStarted ReencryptionProgressEventType = iota
	ReencryptionProgressRunning
	ReencryptionProgressCompleted
	ReencryptionProgressError
)

func (r ReencryptionStatus) String() string {
	switch r {
	case ReencryptionStatusNone:
		return "none"
	case ReencryptionStatusInitialized:
		return "initialized"
	default:
		return fmt.Sprintf("unknown-%d", int(r))
	}
}

func (r ReencryptionProgressEventType) String() string {
	switch r {
	case ReencryptionProgressStarted:
		return "started"
	case ReencryptionProgressRunning:
		return "running"
	case ReencryptionProgressCompleted:
		return "completed"
	case ReencryptionProgressError:
		return "error"
	default:
		return fmt.Sprintf("unknown-%d", int(r))
	}
}

type ReencryptionProgressEvent struct {
	Type                   ReencryptionProgressEventType
	BytesReencryptedSoFar  uint64
	BytesRemaining         uint64
	CalculatedSpeed        int
	EstimatedTimeRemaining int
	TotalTimeSoFar         uint
	Message                string
}

type Reencryption interface {
	Status() (*ReencryptionStatus, error)
	Initialize(unlockKeys map[string][]byte) error
	Resume(unlockKey []byte) (<-chan ReencryptionProgressEvent, error)
}

func FindActiveVolumeForReencryption(ctx context.Context, activeName string) (Reencryption, error) {
	for name, backend := range storageContainerHandlers {
		reencrypt, err := backend.NewOnlineReencryption(ctx, activeName)
		if err != nil {
			return nil, fmt.Errorf("cannot probe %q backend for active name %q: %w", name, activeName, err)
		}
		if reencrypt != nil {
			return reencrypt, nil
		}
	}

	return nil, ErrReencryptionNotActive
}
