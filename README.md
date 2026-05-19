# secboot

A library for managing TPM-backed encrypted disks.

## Features

- System installation:
    * Do platform pre-install checks
    * Initialize the TPM
    * Seal a LUKS passphrase with the TPM
        * Update sealing policy (typically after platform or software update)

- System boot:
    * Unseal
    * Attempt various unlocking paths (TPM-backed, with PIN, with passphrase, revovery...)

- Do factory reset
- Do re-encryption (upcoming)
- Manage recovery keys of a LUKS container (create, list, delete)
- Manage TPM lockout (set an authValue, reset)
- Compute PCR profiles


## Other useful functions

- Get entropy of a PIN or passphrase
- Access to UEFI variables PK, KEK, Db, Dbx


## Authentication modes

- None (or "simple"?)
- Passphrase
- PIN
- Recovery


## Kernel keyring

Keys (LUKS passphrases) are stored in the Kernel keyring when a LUKS container
gets unlocked. They can subsequently be retrieved (GetDiskUnlockKeyFromKernel).

The Primary Key is stored in the Kernel keyring as well
(GetPrimaryKeyFromKernel).


## LUKS header

- Keyslots ...


## Security assets

