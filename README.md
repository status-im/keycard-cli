# keycard-cli

`keycard` is a command line tool to manage [Status Keycards](https://github.com/status-im/status-keycard).

* [Dependencies](#dependencies)
* [Installation](#installation)
* [Continuous Integration](#continuous-integration)
* CLI Commands
  * [Card info](#card-info)
  * [Keycard applet installation](#keycard-applet-installation)
  * [Card initialization](#card-initialization)
  * [Deleting the applet](#deleting-the-applet)
  * [Keycard shell](#keycard-shell)

## Dependencies

On linux you need to install and run the [pcsc daemon](https://linux.die.net/man/8/pcscd).

## Installation

Download the binary for your platform from the [releases page](https://github.com/status-im/keycard-cli/releases).

## Continuous Integration

Jenkins builds provide:

* [PR Builds](https://ci.status.im/job/status-keycard/job/prs/job/keycard-cli/) - Run only the `test` and `build` targets.
* [Manual Builds](https://ci.status.im/job/status-keycard/job/keycard-cli/) - Create GitHub release draft with binaries for 3 platforms.

Successful PR builds are mandatory.

## Usage

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
### Keycard applet installation

The `install` command will install an applet to the card.
You can download the status `cap` file from the [status-im/status-keycard releases page](https://github.com/status-im/status-keycard/releases).

```bash
keycard install -l debug -a PATH_TO_CAP_FILE
```

In case the applet is already installed and you want to force a new installation you can pass the `-f` flag.


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

### Deleting the applet

:warning: **WARNING! This command will remove the applet and all the keys from the card.** :warning:

```bash
keycard-cli delete -l debug
```

### Keycard shell

The shell can be used to interact with the KeyCard using `keycard-go`. You can start the shell with:

```
keycard-cli shell
```

Once in the shell, you may submit one command at a time, followed by Enter.

### Pairing

Before pairing, you need to initialize your card. You can do this interactively in the shell or using `keycard init` specified in the above section.

> Once you initialize your card, **save the PIN, PUK, and Pairing Password fields that are generated**; you will need these to pair with the card. These secrets cannot be set once the card is initialized. If you lose them, you will need to delete the app and reinstall it using the instructions in previous sections.

With the secrets in hand, run the following commands in the shell:

```
> keycard-select
> keycard-set-secrets <PIN> <PUK> <PairingPassword>
> keycard-pair
```

If you don't get an error message, it means you have paired with the card!

