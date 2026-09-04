package luks2

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/snapcore/secboot"
	luks2 "github.com/snapcore/secboot/internal/luks2"
	//"github.com/snapcore/secboot/internal/luksview"
	"github.com/snapcore/secboot/log"
	"io"
	"os/exec"
)

// reencryptionImpl is an implementation of [secboot.Reencryption].
type reencryptionImpl struct {
	sourcePath   string
	dmActiveName string
	ctx          context.Context
}

func (r reencryptionImpl) Status() (*secboot.ReencryptionStatus, error) {
	status, err := luks2.ReadCryptsetupStatus(r.dmActiveName)
	if err != nil {
		return nil, fmt.Errorf("Cannot get reencryption status: %w", err)
	}
	var reencStatus secboot.ReencryptionStatus
	if status.Reencryption == "in-progress" {
		reencStatus = secboot.ReencryptionStatusInitialized
	} else {
		reencStatus = secboot.ReencryptionStatusNone
	}
	return &reencStatus, nil
}

func (r reencryptionImpl) Initialize(ctx context.Context, unlockKeys map[string][]byte) error {
	var unlockKeysOrdered [][]byte
	// TODO: keys must be sorted by their keyslot index
	for keyslotName, unlockKey := range unlockKeys {
		log.Debug("Initialize: keyslotName=%v, unlockKey=%v", keyslotName, unlockKey)
		unlockKeysOrdered = append(unlockKeysOrdered, unlockKey)
	}
	return luks2.ReencryptInitialize(ctx, r.dmActiveName, unlockKeysOrdered)
}

func decodeReencryptionProgressDetails(bytes []byte) (secboot.ReencryptionProgressDetails, error) {
	//log.Debug("decodeReencryptionProgressDetails: %v", string(bytes))
	if len(bytes) == 0 {
		return secboot.ReencryptionProgressDetails{}, fmt.Errorf("empty")
	}
	if bytes[0] != '{' {
		// This is not JSON but probably an error message.
		return secboot.ReencryptionProgressDetails{}, fmt.Errorf("cryptstup error: %v", string(bytes))
	}
	var details secboot.ReencryptionProgressDetails
	err := json.Unmarshal(bytes, &details)
	if err != nil {
		return secboot.ReencryptionProgressDetails{}, fmt.Errorf("cannot decode JSON: %w (%v)", err, string(bytes))
	}
	return details, nil
}

func superviseReencryption(cmd *exec.Cmd, stdoutPipe io.ReadCloser, stderrPipe io.ReadCloser, outChan chan<- secboot.ReencryptionProgressEvent) {
	outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressStarted}

	// Helper function to scan a pipe line by line
	streamLines := func(pipe io.ReadCloser, outputDone chan<- struct{}) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			rawBytes := scanner.Bytes() // should be in JSON format
			details, err := decodeReencryptionProgressDetails(rawBytes)
			if err != nil {
				outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressRunning, Error: err}
			} else {
				outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressRunning, Details: details}
			}
		}
		err := scanner.Err()
		if err != nil {
			outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressRunning, Error: err}
		}
		outputDone <- struct{}{}
	}

	// Read stdout and stderr concurrently
	outputDone := make(chan struct{}, 2)
	go streamLines(stdoutPipe, outputDone)
	go streamLines(stderrPipe, outputDone)

	// Wait for both readers to finish digesting the streams
	<-outputDone
	<-outputDone

	// Wait for the command to clean up and grab the exit status
	err := cmd.Wait()
	if err != nil {
		outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressError}
	} else {
		outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressCompleted}
	}
}

func (r reencryptionImpl) Resume(ctx context.Context, unlockKey []byte) (<-chan secboot.ReencryptionProgressEvent, error) {
	outChan := make(chan secboot.ReencryptionProgressEvent)

	cmd, stdout, stderr, err := luks2.ReencryptResume(ctx, r.dmActiveName, unlockKey)
	if err != nil {
		return nil, err
	}

	go superviseReencryption(cmd, stdout, stderr, outChan)

	return outChan, nil
}
