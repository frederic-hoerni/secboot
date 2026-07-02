package main

import (
	"context"
	"fmt"
	"github.com/snapcore/secboot"
	_ "github.com/snapcore/secboot/luks2" // This gets the LUKS2 backend initialized
	"os"
)

const Usage = `usage: reencrypt [options] DISK

Reencrypt an active encrypted container.

Options:
    --status      Only query the reencryption status
    --initialize  Only initialize reencryption
    --resume      Only resume reencryption

Examples:
    reencrypt --status /dev/sda3
    reencrypt --initialize /dev/sda3
    reencrypt --resume /dev/sda3
`

func main() {
	args := os.Args
	if len(args) <= 1 {
		fmt.Print(Usage)
		os.Exit(1)
	}

	args = args[1:] // pop argument zero
	arg := &args[0]
	args = args[1:] // pop argument zero
	subcommand := reencrypt
	switch *arg {
	case "--status":
		subcommand = status
		arg = nil
	case "--initialize":
		subcommand = initialize
		arg = nil
	case "--resume":
		subcommand = resume
		arg = nil
	default:
	}
	if arg == nil {
		// get it from the next CLI argument
		if len(os.Args) <= 0 {
			fmt.Fprintf(os.Stderr, "Missing DISK argument")
			os.Exit(1)
		}
		arg = &args[0]
	}

	err := subcommand(*arg)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func reencrypt(devicePath string) error {

	err := status(devicePath)
	if err != nil {
		fmt.Printf("Cannot get status of %v: %v\n", devicePath, err)
		return err
	}

	err = initialize(devicePath)
	if err != nil {
		fmt.Printf("Cannot initialize reencryption of %v: %v\n", devicePath, err)
		return err
	}

	err = resume(devicePath)
	if err != nil {
		fmt.Printf("Cannot resume reencryption of %v: %v\n", devicePath, err)
		return err
	}
	return nil
}

func status(devicePath string) error {
	container, err := secboot.FindStorageContainer(context.Background(), devicePath)
	if err != nil {
		fmt.Printf("Cannot find storage (LUKS) container: %v\n", err)
		return err
	}
	reencryption := container.NewReencryption(context.Background())
	status, err := reencryption.Status()
	if err != nil {
		return fmt.Errorf("Cannot get reencryption status: %w", err)
	}
	fmt.Println("Status:", status)
	return nil
}

func initialize(devicePath string) error {
	container, err := secboot.FindStorageContainer(context.Background(), devicePath)
	if err != nil {
		fmt.Printf("Cannot find storage (LUKS) container: %v\n", err)
		return err
	}
	reencryption := container.NewReencryption(context.Background())

	unlockKeys := make(map[string][]byte)
	unlockKeys["1"] = []byte("0000")
	err = reencryption.Initialize(unlockKeys)
	if err != nil {
		return fmt.Errorf("cannot initialize: %w", err)
	}
	fmt.Println("Initialize: ok")
	return nil
}

func resume(devicePath string) error {
	container, err := secboot.FindStorageContainer(context.Background(), devicePath)
	if err != nil {
		fmt.Printf("Cannot find storage (LUKS) container: %v\n", err)
		return err
	}
	reencryption := container.NewReencryption(context.Background())

	reencProgressChannel, err := reencryption.Resume([]byte("0000"))
	if err != nil {
		return fmt.Errorf("cannot resume: %w", err)
	}

	for msg := range reencProgressChannel {
		fmt.Println("xfh: msg=", msg)
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
