package main

import (
	"fmt"
	"github.com/snapcore/secboot"
	"os"
)

func main() {
	fmt.Println("GetDiskUnlockKeyFromKernel")
	unlockKey, err := secboot.GetDiskUnlockKeyFromKernel("ubuntu-fde", os.Args[1], false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Stdout.Write(unlockKey)
	fmt.Println("---")
}
