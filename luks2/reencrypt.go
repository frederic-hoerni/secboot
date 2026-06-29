package luks2

import (
	"bufio"
	"context"
	"fmt"
	"github.com/snapcore/secboot"
	luks2 "github.com/snapcore/secboot/internal/luks2"
	"github.com/snapcore/secboot/internal/luksview"
	"io"
	"os/exec"
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

func superviseReencryption(cmd *exec.Cmd, stdoutPipe io.ReadCloser, stderrPipe io.ReadCloser, outChan chan<- secboot.ReencryptionProgressEvent) {
	outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressStarted}

	// Helper function to scan a pipe line by line
	streamLines := func(pipe io.ReadCloser, outputDone chan<- struct{}) {
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			outChan <- secboot.ReencryptionProgressEvent{Type: secboot.ReencryptionProgressRunning, Message: scanner.Text()}
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

func (r reencryptionImpl) Resume(unlockKey []byte) (<-chan secboot.ReencryptionProgressEvent, error) {
	outChan := make(chan secboot.ReencryptionProgressEvent)

	cmd, stdout, stderr, err := luks2.ReencryptResume(r.path, unlockKey)
	if err != nil {
		return nil, err
	}

	go superviseReencryption(cmd, stdout, stderr, outChan)

	return outChan, nil
}
