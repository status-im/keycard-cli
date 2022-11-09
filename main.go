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
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
)

var version string

type commandFunc func(*scard.Card) error

var (
	logger = log.New("package", "keycard-cli")

	commands map[string]commandFunc
	command  string

	flagCapFile       = flag.String("a", "", "applet cap file path")
	flagKeycardApplet = flag.Bool("keycard-applet", true, "install keycard applet")
	flagCashApplet    = flag.Bool("cash-applet", true, "install cash applet")
	flagNDEFApplet    = flag.Bool("ndef-applet", true, "install NDEF applet")
	flagOverwrite     = flag.Bool("f", false, "force applet installation if already installed")
	flagLogLevel      = flag.String("l", "", `Log level, one of: "error", "warn", "info", "debug", and "trace"`)
	flagNDEFTemplate  = flag.String("ndef", "", "Specify a URL to use in the NDEF record. Use the {{.cashAddress}} variable to get the cash address: http://example.com/{{.cashAddress}}.")
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
		"version": commandVersion,
		"install": commandInstall,
		"info":    commandInfo,
		"delete":  commandDelete,
		"init":    commandInit,
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

func waitForCard(ctx *scard.Context, readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))

	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.StateUnaware
	}

	for {
		for i := range rs {
			if rs[i].EventState&scard.StatePresent != 0 {
				return i, nil
			}

			rs[i].CurrentState = rs[i].EventState
		}

		err := ctx.GetStatusChange(rs, -1)
		if err != nil {
			return -1, err
		}
	}
}

func main() {
	if command == "version" {
		commandVersion(nil)
		return
	}

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

	logger.Info("waiting for a card")
	if len(readers) == 0 {
		fail("no smartcard reader found")
	}

	index, err := waitForCard(ctx, readers)
	if err != nil {
		fail("error waiting for card", "error", err)
	}

	logger.Info("card found", "index", index)
	reader := readers[index]

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

func commandVersion(card *scard.Card) error {
	fmt.Printf("version %+v\n", version)
	return nil
}

func commandInstall(card *scard.Card) error {
	if *flagCapFile == "" {
		logger.Error("you must specify a cap file path with the -a flag\n")
		usage()
	}

	f, err := os.Open(*flagCapFile)
	if err != nil {
		fail("error opening cap file", "error", err)
	}
	defer f.Close()

	i := NewInstaller(card)
	return i.Install(f, *flagOverwrite, *flagKeycardApplet, *flagCashApplet, *flagNDEFApplet, *flagNDEFTemplate)
}

func commandInfo(card *scard.Card) error {
	i := NewInitializer(card)
	info, cashInfo, err := i.Info()
	if err != nil {
		return err
	}

	var keyInitialized bool
	if len(info.KeyUID) > 0 {
		keyInitialized = true
	}

	fmt.Printf("Keycard Applet:\n")
	fmt.Printf("  Installed: %+v\n", info.Installed)
	fmt.Printf("  Initialized: %+v\n", info.Initialized)
	fmt.Printf("  Key Initialized: %+v\n", keyInitialized)
	fmt.Printf("  InstanceUID: 0x%x\n", info.InstanceUID)
	fmt.Printf("  SecureChannelPublicKey: 0x%x\n", info.SecureChannelPublicKey)
	fmt.Printf("  Version: 0x%x\n", info.Version)
	fmt.Printf("  AvailableSlots: 0x%x\n", info.AvailableSlots)
	fmt.Printf("  KeyUID: 0x%x\n", info.KeyUID)
	fmt.Printf("  Capabilities:\n")
	fmt.Printf("    Secure channel:%v\n", info.HasSecureChannelCapability())
	fmt.Printf("    Key management:%v\n", info.HasKeyManagementCapability())
	fmt.Printf("    Credentials Management:%v\n", info.HasCredentialsManagementCapability())
	fmt.Printf("    NDEF:%v\n", info.HasNDEFCapability())
	fmt.Printf("Cash Applet:\n")

	if len(cashInfo.PublicKey) == 0 {
		fmt.Printf("  Installed: %+v\n", false)
		return nil
	}

	ecdsaPubKey, err := crypto.UnmarshalPubkey(cashInfo.PublicKey)
	if err != nil {
		return err
	}

	cashAddress := crypto.PubkeyToAddress(*ecdsaPubKey)

	fmt.Printf("  Installed: %+v\n", cashInfo.Installed)
	fmt.Printf("  PublicKey: 0x%x\n", cashInfo.PublicKey)
	fmt.Printf("  Address: 0x%x\n", cashAddress)
	fmt.Printf("  Public Data: 0x%x\n", cashInfo.PublicData)
	fmt.Printf("  Version: 0x%x\n", cashInfo.Version)

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

func commandShell(card *scard.Card) error {
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		s := NewShell(card)
		return s.Run()
	} else {
		return errors.New("non interactive shell. you must pipe commands")
	}
}
