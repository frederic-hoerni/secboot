package secboot

import "fmt"

type ReencryptionStatus int

const (
	ReencryptionStatusNone = iota
	ReencryptionStatusInterrupted
	ReencryptionStatusInProgress
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
	        case ReencryptionStatusNone: return "none"
		case ReencryptionStatusInterrupted: return "interrupted"
		case ReencryptionStatusInProgress: return "in-progress"
		default: return fmt.Sprintf("unknown-%v", r)
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
