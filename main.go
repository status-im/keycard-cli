package main

import (
	"bufio"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	stdlog "log"
	"os"
	"strconv"
	"strings"

	"github.com/ebfe/scard"
	"github.com/ethereum/go-ethereum/log"
)

type commandFunc func(*scard.Card) error

var (
	logger = log.New("package", "status-go/cmd/keycard")

	commands map[string]commandFunc
	command  string

	flagCapFile   = flag.String("a", "", "applet cap file path")
	flagOverwrite = flag.Bool("f", false, "force applet installation if already installed")
	flagLogLevel  = flag.String("l", "", `Log level, one of: "error", "warn", "info", "debug", and "trace"`)
)

func initLogger() {
	if *flagLogLevel == "" {
		*flagLogLevel = "info"
	}

	level, err := log.LvlFromString(strings.ToLower(*flagLogLevel))
	if err != nil {
		stdlog.Fatal(err)
	}

	handler := log.StreamHandler(os.Stderr, log.TerminalFormat(true))
	filteredHandler := log.LvlFilterHandler(level, handler)
	log.Root().SetHandler(filteredHandler)
}

func init() {
	commands = map[string]commandFunc{
		"install": commandInstall,
		"info":    commandInfo,
		"delete":  commandDelete,
		"init":    commandInit,
		"pair":    commandPair,
		"status":  commandStatus,
		"shell":   commandShell,
	}

	if len(os.Args) < 2 {
		usage()
	}

	command = os.Args[1]
	if len(os.Args) > 2 {
		flag.CommandLine.Parse(os.Args[2:])
	}

	initLogger()
}

func usage() {
	fmt.Printf("\nUsage:\n  keycard COMMAND [FLAGS]\n\nAvailable commands:\n")
	for name := range commands {
		fmt.Printf("  %s\n", name)
	}
	fmt.Print("\nFlags:\n\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func fail(msg string, ctx ...interface{}) {
	logger.Error(msg, ctx...)
	os.Exit(1)
}

func main() {
	ctx, err := scard.EstablishContext()
	if err != nil {
		fail("error establishing card context", "error", err)
	}
	defer func() {
		if err := ctx.Release(); err != nil {
			logger.Error("error releasing context", "error", err)
		}
	}()

	readers, err := ctx.ListReaders()
	if err != nil {
		fail("error getting readers", "error", err)
	}

	if len(readers) == 0 {
		fail("couldn't find any reader")
	}

	if len(readers) > 1 {
		fail("too many readers found")
	}

	reader := readers[0]
	logger.Debug("using reader", "name", reader)
	logger.Debug("connecting to card", "reader", reader)
	card, err := ctx.Connect(reader, scard.ShareShared, scard.ProtocolAny)
	if err != nil {
		fail("error connecting to card", "error", err)
	}
	defer func() {
		if err := card.Disconnect(scard.ResetCard); err != nil {
			logger.Error("error disconnecting card", "error", err)
		}
	}()

	status, err := card.Status()
	if err != nil {
		fail("error getting card status", "error", err)
	}

	switch status.ActiveProtocol {
	case scard.ProtocolT0:
		logger.Debug("card protocol", "T", "0")
	case scard.ProtocolT1:
		logger.Debug("card protocol", "T", "1")
	default:
		logger.Debug("card protocol", "T", "unknown")
	}

	if f, ok := commands[command]; ok {
		err = f(card)
		if err != nil {
			logger.Error("error executing command", "command", command, "error", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	fail("unknown command", "command", command)
	usage()
}

func ask(description string) string {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("%s: ", description)
	text, err := r.ReadString('\n')
	if err != nil {
		stdlog.Fatal(err)
	}

	return strings.TrimSpace(text)
}

func askHex(description string) []byte {
	s := ask(description)
	if s[:2] == "0x" {
		s = s[2:]
	}

	data, err := hex.DecodeString(s)
	if err != nil {
		stdlog.Fatal(err)
	}

	return data
}

func askInt(description string) int {
	s := ask(description)
	i, err := strconv.ParseInt(s, 10, 8)
	if err != nil {
		stdlog.Fatal(err)
	}

	return int(i)
}

func commandInstall(card *scard.Card) error {
	if *flagCapFile == "" {
		logger.Error("you must specify a cap file path with the -f flag\n")
		usage()
	}

	f, err := os.Open(*flagCapFile)
	if err != nil {
		fail("error opening cap file", "error", err)
	}
	defer f.Close()

	i := NewInstaller(card)

	return i.Install(f, *flagOverwrite)
}

func commandInfo(card *scard.Card) error {
	i := NewInitializer(card)
	info, err := i.Info()
	if err != nil {
		return err
	}

	fmt.Printf("Installed: %+v\n", info.Installed)
	fmt.Printf("Initialized: %+v\n", info.Initialized)
	fmt.Printf("InstanceUID: 0x%x\n", info.InstanceUID)
	fmt.Printf("SecureChannelPublicKey: 0x%x\n", info.SecureChannelPublicKey)
	fmt.Printf("Version: 0x%x\n", info.Version)
	fmt.Printf("AvailableSlots: 0x%x\n", info.AvailableSlots)
	fmt.Printf("KeyUID: 0x%x\n", info.KeyUID)
	fmt.Printf("Capabilities:\n")
	fmt.Printf("  Secure channel:%v\n", info.HasSecureChannelCapability())
	fmt.Printf("  Key management:%v\n", info.HasKeyManagementCapability())
	fmt.Printf("  Credentials Management:%v\n", info.HasCredentialsManagementCapability())
	fmt.Printf("  NDEF:%v\n", info.HasNDEFCapability())

	return nil
}

func commandDelete(card *scard.Card) error {
	i := NewInstaller(card)
	err := i.Delete()
	if err != nil {
		return err
	}

	fmt.Printf("applet deleted\n")

	return nil
}

func commandInit(card *scard.Card) error {
	i := NewInitializer(card)
	secrets, err := i.Init()
	if err != nil {
		return err
	}

	fmt.Printf("PIN %s\n", secrets.Pin())
	fmt.Printf("PUK %s\n", secrets.Puk())
	fmt.Printf("Pairing password: %s\n", secrets.PairingPass())

	return nil
}

func commandPair(card *scard.Card) error {
	i := NewInitializer(card)
	pairingPass := ask("Pairing password")
	info, err := i.Pair(pairingPass)
	if err != nil {
		return err
	}

	fmt.Printf("Pairing key 0x%x\n", info.Key)
	fmt.Printf("Pairing Index %d\n", info.Index)

	return nil
}

func commandStatus(card *scard.Card) error {
	i := NewInitializer(card)
	key := askHex("Pairing key")
	index := askInt("Pairing index")

	appStatus, err := i.Status(key, index)
	if err != nil {
		return err
	}

	fmt.Printf("Pin retry count: %d\n", appStatus.PinRetryCount)
	fmt.Printf("PUK retry count: %d\n", appStatus.PUKRetryCount)
	fmt.Printf("Key initialized: %v\n", appStatus.KeyInitialized)

	return nil
}

func commandShell(card *scard.Card) error {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		s := NewShell(card)
		return s.Run()
	} else {
		return errors.New("non interactive shell. you must pipe commands")
	}
}
