package secboot

import (
	"fmt"
)

type ReencryptionStatus int

const (
	ReencryptionStatusNone = iota
	ReencryptionStatusInitialized
)

type ReencryptionProgressEventType int

const (
	ReencryptionProgressStarted   = iota
	ReencryptionProgressRunning   = iota
	ReencryptionProgressCompleted = iota
	ReencryptionProgressError     = iota
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
