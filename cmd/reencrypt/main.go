package main

import (
	"context"
	"fmt"
	efi "github.com/canonical/go-efilib"
	"github.com/snapcore/secboot"
	"github.com/snapcore/secboot/internal/luksview"
	_ "github.com/snapcore/secboot/luks2" // This gets the LUKS2 backend initialized
	"os"
)

const Usage = `usage: reencrypt DISK1 ...
`

func main() {
	//fmt.Println("main: os.Args=", os.Args)
	if len(os.Args) <= 1 {
		fmt.Print(Usage)
		os.Exit(1)
	}
	args := os.Args[1:]
	//_ = test1(args[0])
	err := test2(args[0])
	err = reencrypt(args[0])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}

func reencrypt(devicePath string) error {
	container, err := secboot.FindStorageContainer(efi.DefaultVarContext, devicePath)
	if err != nil {
		fmt.Printf("Cannot find storage (LUKS) container: %v\n", err)
		return err
	}
	reencryption := container.NewReencryption()
	status, err := reencryption.Status()
	if err != nil {
		return fmt.Errorf("cannot get status: %w", err)
	}
	fmt.Println("Status:", *status)

	var unlockKeys map[string][]byte
	err = reencryption.Initialize(unlockKeys)
	if err != nil {
		return fmt.Errorf("cannot initialize: %w", err)
	}
	fmt.Println("Initialize: ok")
	status, err = reencryption.Status()
	fmt.Println("Status:", *status)

	reencProgressChannel, err := reencryption.Resume([]byte("0000"))
	if err != nil {
		return fmt.Errorf("cannot resume: %w", err)
	}

	status, err = reencryption.Status()
	fmt.Println("Status:", *status)

	for msg := range reencProgressChannel {
		fmt.Println("xfh: msg=", msg)
	}
	return nil
}

func test1(devicePath string) error {
	container, err := secboot.FindStorageContainer(efi.DefaultVarContext, devicePath)
	if err != nil {
		fmt.Printf("Cannot find storage (LUKS) container: %v\n", err)
		return err
	}
	fmt.Println("container:", container, "backed:", container.BackendName())

	reencryption := container.NewReencryption()

	fmt.Println("reencryption:", reencryption)

	return nil
}

func test2(devicePath string) error {
	view, err := luksview.NewView(context.TODO(), devicePath)
	if err != nil {
		return fmt.Errorf("cannot obtain LUKS header view: %w", err)
	}
	nReencrypt := view.HasReencrypt()
	fmt.Println("test2: nReencrypt=", nReencrypt)
	return nil
}
