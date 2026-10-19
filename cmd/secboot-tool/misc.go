package main

import (
	"context"
	"fmt"
	"github.com/snapcore/secboot"
	"github.com/snapcore/secboot/internal/luksview"
	"github.com/snapcore/secboot/tpm2"
	"os"
)

func cmdGetDiskUnlockKeyFromKernel(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("bad argument count\nusage: seboot-tool get-unlockkey-from-kernel <device-path>")
	}
	devicePath := args[0]
	unlockKey, err := secboot.GetDiskUnlockKeyFromKernel("ubuntu-fde", devicePath, false)
	if err != nil {
		return err
	}
	os.Stdout.Write(unlockKey)
	return nil
}

func cmdPrintTokens(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("bad argument count\nusage: seboot-tool print-tokens <device-path>")
	}

	devicePath := args[0]

	view, err := luksview.NewView(context.Background(), devicePath)
	if err != nil {
		return err
	}
	format := "%-30v %-30v\n"
	fmt.Printf(format, "NAME", "TYPE")
	for _, name := range view.TokenNames() {
		token, _, _ := view.TokenByName(name)
		fmt.Printf(format, name, token.Type())
	}
	return nil
}

func cmdPrintSealedData(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("Missing argument: device path")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}
	devicePath := args[0]
	container, err := secboot.FindStorageContainer(context.Background(), devicePath)
	if err != nil {
		return err
	}

	reader, err := container.OpenRead(context.Background())
	if err != nil {
		return err
	}
	defer reader.Close()

	names, err := reader.ListKeyslotNames(context.Background())
	if err != nil {
		return err
	}

	for _, name := range names {
		slot, err := reader.ReadKeyslot(context.Background(), name)
		if err != nil {
			fmt.Println("SealedKeyData for keyslot:", name)
			return err
		}

		fmt.Printf("SealedKeyData for keyslot '%v' (type: %v)\n", name, slot.Type())
		if slot.Type() != secboot.KeyslotTypePlatform {
			fmt.Printf("Type not '%v', ignore...\n", secboot.KeyslotTypePlatform)
			continue
		}
		kd, err := secboot.ReadKeyData(slot.Data())
		if err != nil {
			return err
		}

		sealedKeyData, err := tpm2.NewSealedKeyData(kd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot print sealed object for '%v': %v\n", err.Error(), name)
			continue
		}
		fmt.Println(sealedKeyData.String("   "))
	}

	return nil
}
