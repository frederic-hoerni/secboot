package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"github.com/snapcore/secboot"
	"github.com/snapcore/secboot/log"
	_ "github.com/snapcore/secboot/luks2" // This gets the LUKS2 backend initialized
	"os"
)

const Usage = `
usage: 1. reencrypt              ACTIVE-NAME KEYSLOT-NAME:UNLOCK-KEY-HEX ...
       2. reencrypt --status     ACTIVE-NAME
       3. reencrypt --initialize ACTIVE-NAME KEYSLOT-NAME:UNLOCK-KEY-HEX ...
       4. reencrypt --resume     ACTIVE-NAME UNLOCK-KEY-HEX

Reencrypt an active encrypted container.

Options:
  --status      Only query the reencryption status
  --initialize  Only initialize reencryption
  --resume      Only resume reencryption

For initializing reencryption, all unlock keys must be provided with
the name of their keyslot.

Arguments:
  ACTIVE-NAME     Name of the active volume as shown in /dev/mapper
  KEYSLOT-NAME    Name of the keyslot, as managed by secboot
  UNLOCK-KEY-HEX  Unlock key (hexadecimal)

Examples:
  cryptsetup open /dev/vda5 crypt01
  reencrypt --status crypt01
  reencrypt --initialize crypt01 default-recovery:010203... default:00112233...
  reencrypt --resume crypt01 010203...
`

func main() {
	args := os.Args
	if len(args) <= 1 {
		fmt.Print(Usage)
		os.Exit(1)
	}

	log.SetLogLevel(log.LogLevelInfo)

	args = args[1:] // pop argument zero
	arg := &args[0]
	var err error
	switch *arg {
	case "--status":
		err = status(args[1:]...)
	case "--initialize":
		err = initialize(args[1:]...)
	case "--resume":
		err = resume(args[1:]...)
	default:
		err = reencrypt(args...)
	}

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

// Split "KEY:VALUE" into "KEY" and "VALUE"
func splitUnlockKeySpec(s string) (string, string, error) {
	items := strings.Split(s, ":")
	if len(items) != 2 {
		return "", "", fmt.Errorf("invalid format KEY:VALUE")
	}
	return items[0], items[1], nil
}

//func reencrypt(activeName string, unlockKeysHex []string) error {
func reencrypt(args ...string) error {

	if len(args) < 2 {
		return fmt.Errorf("reencrypt: not enough arguments")
	}
	activeName := args[0]
	unlockKeysHex := args[1:] // remaining arguments

	log.Info("reencrypt %v with %v key(s)", activeName, len(unlockKeysHex))

	err := status(activeName)
	if err != nil {
		return err
	}

	err = initialize(append([]string{activeName}, unlockKeysHex...)...)
	if err != nil {
		fmt.Printf("Cannot initialize reencryption of %v: %v\n", activeName, err)
		return err
	}

	// Resume using the first key of the list
	_, unlockKey, err := splitUnlockKeySpec(unlockKeysHex[0])
	if err != nil {
		// Should not happen, as already validated in initialize above
		return err
	}
	err = resume(activeName, unlockKey)
	if err != nil {
		fmt.Printf("Cannot resume reencryption of %v: %v\n", activeName, err)
		return err
	}
	return nil
}

func status(args ...string) error {
	if len(args) != 1 {
		return fmt.Errorf("status: bad argument count")
	}
	activeName := args[0]

	log.Info("reencrypt status %v", activeName)

	reencryption, err := secboot.ReencryptionForActiveVolume(activeName)
	if err != nil {
		return err
	}

	status, err := reencryption.Status()
	if err != nil {
		return err
	}
	fmt.Println("Status:", status)
	return nil
}

func initialize(args ...string) error {
	if len(args) < 2  {
		return fmt.Errorf("initialize: missing argument")
	}

	activeName := args[0]
	unlockKeysHex := args[1:] // remaining arguments

	log.Info("reencrypt initialize %v", activeName)

	reencryption, err := secboot.ReencryptionForActiveVolume(activeName)
	if err != nil {
		return fmt.Errorf("Cannot find active volume: %w\n", err)
	}

	unlockKeys := make(map[string][]byte)
	for _, unlockKeySpec := range unlockKeysHex {
		// split name:hex into a map
		key, unlockKeyHex, err := splitUnlockKeySpec(unlockKeySpec)
		if err != nil {
			return err
		}

		unlockKey, err := hex.DecodeString(unlockKeyHex)
		if err != nil {
			return fmt.Errorf("Malformed hex '%v': %w", unlockKeyHex, err)
		}
		unlockKeys[key] = unlockKey
	}

	err = reencryption.Initialize(context.Background(), unlockKeys)
	if err != nil {
		return fmt.Errorf("Cannot initialize: %w", err)
	}
	fmt.Println("Initialize: ok")
	return nil
}

func resume(args ...string) error {
	if len(args) != 2  {
		return fmt.Errorf("initialize: bad argument count")
	}

	activeName := args[0]
	unlockKeyHex := args[1]

	log.Info("reencrypt resume %v", activeName)
	unlockKey, err := hex.DecodeString(unlockKeyHex)
	if err != nil {
		return fmt.Errorf("Malformed hex '%v': %w", unlockKeyHex, err)
	}

	reencryption, err := secboot.ReencryptionForActiveVolume(activeName)
	if err != nil {
		fmt.Printf("Cannot find active volume: %v\n", err)
		return err
	}

	reencProgressChannel, err := reencryption.Resume(context.Background(), unlockKey)
	if err != nil {
		return fmt.Errorf("cannot resume: %w", err)
	}

	i := 0

	for msg := range reencProgressChannel {
		fmt.Println("xfh: i=", i, "msg=", msg)
		i++
		switch msg.Type {
		case secboot.ReencryptionProgressCompleted:
			return nil
		case secboot.ReencryptionProgressError:
			return nil
		case secboot.ReencryptionProgressStarted:
			continue
		case secboot.ReencryptionProgressRunning:
			continue
		}
	}
	return nil
}
