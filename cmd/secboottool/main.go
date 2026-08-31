package main

import (
	"fmt"
	"github.com/snapcore/secboot"
	"os"
)

const Usage = `
usage: 1. secboottool GetDiskUnlockKeyFromKernel DEVICE
       2. secboottool ListLUKS2ContainerKeyNames DEVICE
`

func usage() {
	fmt.Print(Usage)
	os.Exit(1)
}

func main() {

	args := os.Args
	if len(args) <= 1 {
		usage()
	}

	arg := args[1]
	args = args[2:] // pop 2 first items from the command line
	switch arg {
	case "GetDiskUnlockKeyFromKernel":
		if len(args) == 0 {
			usage()
		}
		GetDiskUnlockKeyFromKernel(args[0])
	case "ListLUKS2ContainerKeyNames":
		if len(args) == 0 {
			usage()
		}
		ListLUKS2ContainerKeyNames(args[0])
	default:
		usage()
	}
}

func GetDiskUnlockKeyFromKernel(devicePath string) {
	fmt.Println("GetDiskUnlockKeyFromKernel")
	unlockKey, err := secboot.GetDiskUnlockKeyFromKernel("ubuntu-fde", devicePath, false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Stdout.Write(unlockKey)
	fmt.Println("---")
}

func ListLUKS2ContainerKeyNames(devicePath string) {
	fmt.Println("ListLUKS2ContainerKeyNames")
	names, err := secboot.ListLUKS2ContainerUnlockKeyNames(devicePath)
	if err != nil {
		fmt.Printf("ListLUKS2ContainerUnlockKeyNames -> err=%v\n", err)
	} else {
		fmt.Printf("ListLUKS2ContainerUnlockKeyNames -> %v\n", names)
	}
	names, err = secboot.ListLUKS2ContainerRecoveryKeyNames(devicePath)
	if err != nil {
		fmt.Printf("ListLUKS2ContainerRecoveryKeyNames -> err=%v\n", err)
	} else {
		fmt.Printf("ListLUKS2ContainerRecoveryKeyNames -> %v\n", names)
	}

}
