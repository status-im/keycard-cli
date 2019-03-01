# status-hardware-wallet

`status-hardware-wallet` is a command line tool you can use to initialize a smartcard with the [Status Hardware Wallet](https://github.com/status-im/hardware-wallet).

## Dependencies

To install `hardware-wallet-go` you need `go` in your system.

MacOSX:

`brew install go`

## Installation

`go get github.com/status-im/hardware-wallet-go/cmd/status-hardware-wallet`

The executable will be installed in `$GOPATH/bin`.
Check your `$GOPATH` with `go env`.

## Usage

### Install the hardware wallet applet

The install command will install an applet to the card.
You can download the status `cap` file from the [status-im/hardware-wallet releases page](https://github.com/status-im/hardware-wallet/releases).

```bash
status-hardware-wallet install -l debug -a PATH_TO_CAP_FILE
```

In case the applet is already installed and you want to force a new installation you can pass the `-f` flag.

### Card info

```bash
status-hardware-wallet info -l debug
```

The `info` command will print something like this:

```
Installed: true
Initialized: false
InstanceUID: 0x
PublicKey: 0x112233...
Version: 0x
AvailableSlots: 0x
KeyUID: 0x
```

### Card initialization


```bash
status-hardware-wallet init -l debug
```

The `init` command initializes the card and generates the secrets needed to pair the card to a device.

```
PIN 123456
PUK 123456789012
Pairing password: RandomPairingPassword
```

### Deleting the applet from the card

:warning: **WARNING! This command will remove the applet and all the keys from the card.** :warning:

```bash
status-hardware-wallet delete -l debug
```

### Pairing

```bash
status-hardware-wallet pair -l debug
```

The process will ask for `PairingPassword` and `PIN` and will generate a pairing key you can use to interact with the card.
