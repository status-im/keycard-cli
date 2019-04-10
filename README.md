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

### Talking to the Card

Before you can communicate with the card, you need to initialize it. You can do this interactively in the shell or using `keycard init` specified in the above section.

> Once you initialize your card, **save the PIN, PUK, and Pairing Password fields that are generated**; you will need these to establish a connection with the card.

With the secrets in hand, run the following commands in the shell:

```
> keycard-select
> keycard-set-secrets <PIN> <PUK> <PairingPassword>
> keycard-pair
> keycard-open-secure-channel
> keycard-verify-pin <PIN>
```

If you don't get an error message, it means you are connected to the card! This connection will persist for the duration of your shell session.

> `keycard-pair` will print two values: `PAIRING KEY` and `PAIRING INDEX`. You should save these values if you wish to reuse this pairing with a new shell session. You can do this with the command: `>keycard-set-pairing <PAIRING KEY> <PAIRING INDEX>`

### Wallets

Once comms are established with the card, you should generate a keypair on the card if you haven't done so already:

```
> keycard-generate-key
```

This will create a keypair that does not correspond to a mnemonic, so you cannot export it as a seed phrase. The generation utilizes the card's true random number generator.

> The KeyCard applet does have the ability to import a key with a menomonic, but that isn't as secure as generating on the card. It is also not implemented in this CLI or in the Golang SDK yet.

#### Setting a "Current" Key

The card applet works by loading a private key (based on a derivation path) into a "current" state. Once set as the current key, the subsequent signature request will be filled by that key.

You can make a key current by running:

```
> keycard-derive-key <DERIVATION_PATH>
```

Where `DERIVATION_PATH` is a BIP44 path, e.g. `m/44/0'/0'/0/0`

#### Exporting Keys

You can export both public and private keys based on a derivation path (or without a path, using the current key).

> Public keys can always be exported, but private keys have some restrictions, which you can read about [here].

When exporting a key, you have three parameters to specify:

* `derive` - if false, the card will just return the current key. This means you must have derived (i.e. made current) the key ahead of time
* `makeCurrent` - if true, the card will make the derived key "current" before returning it
* `onlyPublic` - If true, only return the f false, return the private and public key (again, there are restrictions)
* `path` - Derivation path

Example:

```
> keycard-export-key 1 0 0 m/44'/0'/0'/0/0

<RESPONSE_KEY>
```

Where `RESPONSE_KEY` is an EC public key on the secp256k1 curve in uncompressed point format, i.e. `04{X-component}{Y-component}`.
