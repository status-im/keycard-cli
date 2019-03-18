# keycard-cli

`keycard-cli` is a command line tool to manage [Status Keycards](https://github.com/status-im/status-keycard).

## Dependencies

To install `keycard-go` you need `go` in your system.

MacOSX:

`brew install go`

## Installation

`go get -u github.com/status-im/keycard-cli`

The executable will be installed in `$GOPATH/bin`.
Check your `$GOPATH` with `go env`.

## Usage

### Install the keycard applet

The install command will install an applet to the card.
You can download the status `cap` file from the [status-im/status-keycard releases page](https://github.com/status-im/status-keycard/releases).

```bash
keycard install -l debug -a PATH_TO_CAP_FILE
```

In case the applet is already installed and you want to force a new installation you can pass the `-f` flag.

### Card info

```bash
keycard info -l debug
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
keycard init -l debug
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
keycard delete -l debug
```
