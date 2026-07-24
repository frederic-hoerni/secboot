package main

import (
	"fmt"
	"github.com/canonical/go-tpm2"
	"github.com/snapcore/secboot/efi"
	secboot_tpm2 "github.com/snapcore/secboot/tpm2"
	"os"
)

const Usage = `usage: add-pcr-profile [<options>] IMAGE ...

This computes and prints a PCR profile for the given sequence of images.

OPTIONS
    --event-log TCG-EVENT-LOG
    --efivars EFI-VARS (TODO)
        default: /sys/firmware/efi/efivars
`

func main() {
	args := os.Args
	//if len(args) != 2 {
	//	fmt.Print(Usage)
	//	os.Exit(1)
	//}

	// TODO: Get arg --event-log
	//tcgLogFile := args[1]
	//efi.SetEventLogPath(tcgLogFile)
	//log, err := efi.DefaultEnv.ReadEventLog()
	//if err != nil {
	//	fmt.Println("error:", err)
	//	os.Exit(1)
	//}
	//fmt.Println("log=", log)

	profile := secboot_tpm2.NewPCRProtectionProfile()

	sequences := efi.NewImageLoadSequences()
	args = args[1:]
	var imageLoadActivity efi.ImageLoadActivity
	for index, arg := range args {
		fmt.Println("image:", index, arg)
		if index == 0 {
			imageLoadActivity = efi.NewImageLoadActivity(efi.FileImage(arg))
		} else {
			imageLoadActivity.Loads(efi.NewImageLoadActivity(efi.FileImage(arg)))
		}
	}
	sequences = sequences.Append(imageLoadActivity)

	err := efi.AddPCRProfile(
		tpm2.HashAlgorithmSHA256,
		profile.RootBranch(),
		sequences,
		efi.WithKernelConfigProfile(),
		efi.WithSecureBootPolicyProfile(),
	)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Profile:", profile)
	}
}
