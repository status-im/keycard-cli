package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	keycard "github.com/status-im/keycard-go"
	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/globalplatform"
	"github.com/status-im/keycard-go/types"
)

type shellCommand = func(args ...string) error

type Shell struct {
	t        globalplatform.Transmitter
	c        types.Channel
	secrets  *keycard.Secrets
	gpCmdSet *globalplatform.CommandSet
	kCmdSet  *keycard.CommandSet
	commands map[string]shellCommand
	out      *bytes.Buffer
}

func NewShell(t globalplatform.Transmitter) *Shell {
	c := globalplatform.NewNormalChannel(t)

	s := &Shell{
		t:        t,
		c:        c,
		kCmdSet:  keycard.NewCommandSet(c),
		gpCmdSet: globalplatform.NewCommandSet(c),
		out:      new(bytes.Buffer),
	}

	s.commands = map[string]shellCommand{
		"gp-send-apdu":                s.commandGPSendAPDU,
		"gp-select":                   s.commandGPSelect,
		"gp-open-secure-channel":      s.commandGPOpenSecureChannel,
		"gp-delete":                   s.commandGPDelete,
		"gp-load":                     s.commandGPLoad,
		"gp-install-for-install":      s.commandGPInstallForInstall,
		"keycard-init":                s.commandKeycardInit,
		"keycard-select":              s.commandKeycardSelect,
		"keycard-pair":                s.commandKeycardPair,
		"keycard-open-secure-channel": s.commandKeycardOpenSecureChannel,
		"keycard-get-status":          s.commandKeycardGetStatus,
		"keycard-set-secrets":         s.commandKeycardSetSecrets,
		"keycard-set-pairing":         s.commandKeycardSetPairing,
		"keycard-verify-pin":          s.commandKeycardVerifyPIN,
		"keycard-generate-key":        s.commandKeycardGenerateKey,
		"keycard-derive-key":          s.commandKeycardDeriveKey,
		"keycard-sign":                s.commandKeycardSign,
		"keycard-set-pinless-path":    s.commandKeycardSetPinlessPath,
	}

	return s
}

func (s *Shell) write(str string) {
	s.out.WriteString(str)
}

func (s *Shell) flushOut() {
	io.Copy(os.Stdout, s.out)
}

func (s *Shell) Run() error {
	reader := bufio.NewReader(os.Stdin)
	defer s.flushOut()

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		err = s.evalLine(line)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Shell) commandGPSendAPDU(args ...string) error {
	if err := s.requireArgs(args, 1); err != nil {
		return err
	}

	apdu, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("send apdu %x", apdu))
	resp, err := s.t.Transmit(apdu)
	if err != nil {
		logger.Error("send apdu failed", "error", err)
		return err
	}
	logger.Info(fmt.Sprintf("raw response: %x", resp))

	return nil
}

func (s *Shell) commandGPSelect(args ...string) error {
	if err := s.requireArgs(args, 0, 1); err != nil {
		return err
	}

	if len(args) == 0 {
		logger.Info("select ISD")
		return s.gpCmdSet.Select()
	}

	aid, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("select AID %x", aid))
	return s.gpCmdSet.SelectAID(aid)
}

func (s *Shell) commandGPOpenSecureChannel(args ...string) error {
	if err := s.requireArgs(args, 0); err != nil {
		return err
	}

	logger.Info("open secure channel")
	return s.gpCmdSet.OpenSecureChannel()
}

func (s *Shell) commandGPDelete(args ...string) error {
	if err := s.requireArgs(args, 1); err != nil {
		return err
	}

	aid, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("delete %x", aid))

	return s.gpCmdSet.Delete(aid)
}

func (s *Shell) commandGPLoad(args ...string) error {
	if err := s.requireArgs(args, 2); err != nil {
		return err
	}

	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()

	pkgAID, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}

	logger.Info("load package")
	callback := func(index, total int) {
		logger.Debug(fmt.Sprintf("loading %d/%d", index+1, total))
	}
	if err = s.gpCmdSet.LoadPackage(f, pkgAID, callback); err != nil {
		logger.Error("load failed", "error", err)
		return err
	}

	return nil
}

func (s *Shell) commandGPInstallForInstall(args ...string) error {
	if err := s.requireArgs(args, 3, 4); err != nil {
		return err
	}

	pkgAID, err := hex.DecodeString(args[0])
	if err != nil {
		return err
	}

	appletAID, err := hex.DecodeString(args[1])
	if err != nil {
		return err
	}

	instanceAID, err := hex.DecodeString(args[2])
	if err != nil {
		return err
	}

	var params []byte

	if len(args) == 4 {
		params, err = hex.DecodeString(args[3])
		if err != nil {
			return err
		}
	}

	logger.Info("install for install", "pkg", pkgAID, "applet", appletAID, "instance", instanceAID, "params", params)

	return s.gpCmdSet.InstallForInstall(pkgAID, appletAID, instanceAID, params)
}

func (s *Shell) commandKeycardInit(args ...string) error {
	if err := s.requireArgs(args, 0); err != nil {
		return err
	}

	i := NewInitializer(s.t)
	secrets, err := i.Init()
	if err != nil {
		return err
	}

	s.secrets = secrets
	s.write(fmt.Sprintf("PIN: %s\n", secrets.Pin()))
	s.write(fmt.Sprintf("PUK: %s\n", secrets.Puk()))
	s.write(fmt.Sprintf("PAIRING PASSWORD: %s\n", secrets.PairingPass()))

	return nil
}

func (s *Shell) commandKeycardSetSecrets(args ...string) error {
	if err := s.requireArgs(args, 3); err != nil {
		return err
	}

	s.secrets = keycard.NewSecrets(args[0], args[1], args[2])

	return nil
}

func (s *Shell) commandKeycardSelect(args ...string) error {
	if err := s.requireArgs(args, 0); err != nil {
		return err
	}

	logger.Info("select keycard")
	err := s.kCmdSet.Select()
	info := s.kCmdSet.ApplicationInfo

	s.write(fmt.Sprintf("Installed: %+v\n", info.Installed))
	s.write(fmt.Sprintf("Initialized: %+v\n", info.Initialized))
	s.write(fmt.Sprintf("InstanceUID: %x\n", info.InstanceUID))
	s.write(fmt.Sprintf("SecureChannelPublicKey: %x\n", info.SecureChannelPublicKey))
	s.write(fmt.Sprintf("Version: %x\n", info.Version))
	s.write(fmt.Sprintf("AvailableSlots: %x\n", info.AvailableSlots))
	s.write(fmt.Sprintf("KeyUID: %x\n", info.KeyUID))
	s.write(fmt.Sprintf("Capabilities:\n"))
	s.write(fmt.Sprintf("  Secure channel:%v\n", info.HasSecureChannelCapability()))
	s.write(fmt.Sprintf("  Key management:%v\n", info.HasKeyManagementCapability()))
	s.write(fmt.Sprintf("  Credentials Management:%v\n", info.HasCredentialsManagementCapability()))
	s.write(fmt.Sprintf("  NDEF:%v\n", info.HasNDEFCapability()))

	if e, ok := err.(*apdu.ErrBadResponse); ok && e.Sw == globalplatform.SwFileNotFound {
		logger.Error("select keycard failed", "error", err)
		return ErrNotInstalled
	}

	return err
}

func (s *Shell) commandKeycardPair(args ...string) error {
	if err := s.requireArgs(args, 0); err != nil {
		return err
	}

	if s.secrets == nil {
		return errors.New("cannot pair without setting secrets")
	}

	err := s.kCmdSet.Pair(s.secrets.PairingPass())
	if err != nil {
		return err
	}

	s.write(fmt.Sprintf("PAIRING KEY: %x\n", s.kCmdSet.PairingInfo.Key))
	s.write(fmt.Sprintf("PAIRING INDEX: %v\n", s.kCmdSet.PairingInfo.Index))

	return nil
}

func (s *Shell) commandKeycardSetPairing(args ...string) error {
	if err := s.requireArgs(args, 2); err != nil {
		return err
	}

	key, err := s.parseHex(args[0])
	if err != nil {
		return err
	}

	index, err := strconv.ParseInt(args[1], 10, 8)
	if err != nil {
		return err
	}

	s.kCmdSet.SetPairingInfo(key, int(index))

	return nil
}

func (s *Shell) commandKeycardOpenSecureChannel(args ...string) error {
	if err := s.requireArgs(args, 0); err != nil {
		return err
	}

	if s.kCmdSet.PairingInfo == nil {
		return errors.New("cannot open secure channel without setting pairing info")
	}

	logger.Info("open keycard secure channel")
	if err := s.kCmdSet.OpenSecureChannel(); err != nil {
		logger.Error("open keycard secure channel failed", "error", err)
		return err
	}

	return nil
}

func (s *Shell) commandKeycardGetStatus(args ...string) error {
	if err := s.requireArgs(args, 0); err != nil {
		return err
	}

	logger.Info("get status application")
	appStatus, err := s.kCmdSet.GetStatusApplication()
	if err != nil {
		logger.Error("get status application failed", "error", err)
		return err
	}

	logger.Info("get status key path")
	keyStatus, err := s.kCmdSet.GetStatusKeyPath()
	if err != nil {
		logger.Error("get status key path failed", "error", err)
		return err
	}

	s.write(fmt.Sprintf("STATUS - PIN RETRY COUNT: %d\n", appStatus.PinRetryCount))
	s.write(fmt.Sprintf("STATUS - PUK RETRY COUNT: %d\n", appStatus.PUKRetryCount))
	s.write(fmt.Sprintf("STATUS - KEY INITIALIZED: %v\n", appStatus.KeyInitialized))
	s.write(fmt.Sprintf("STATUS - KEY PATH: %v\n", keyStatus.Path))

	return nil
}

func (s *Shell) commandKeycardVerifyPIN(args ...string) error {
	if err := s.requireArgs(args, 1); err != nil {
		return err
	}

	logger.Info("verify PIN")
	if err := s.kCmdSet.VerifyPIN(args[0]); err != nil {
		logger.Error("verify PIN failed", "error", err)
		return err
	}

	return nil
}

func (s *Shell) commandKeycardGenerateKey(args ...string) error {
	if err := s.requireArgs(args, 0); err != nil {
		return err
	}

	logger.Info("generate key")
	keyUID, err := s.kCmdSet.GenerateKey()
	if err != nil {
		return err
	}

	s.write(fmt.Sprintf("KEY UID %x\n", keyUID))

	return nil
}

func (s *Shell) commandKeycardDeriveKey(args ...string) error {
	if err := s.requireArgs(args, 1); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("derive key %s", args[0]))
	if err := s.kCmdSet.DeriveKey(args[0]); err != nil {
		logger.Error("derive key failed", "error", err)
		return err
	}

	return nil
}

func (s *Shell) commandKeycardSign(args ...string) error {
	if err := s.requireArgs(args, 1); err != nil {
		return err
	}

	data, err := s.parseHex(args[0])
	if err != nil {
		logger.Error("failed parsing hex data", "error", err)
		return err
	}

	logger.Info("sign")
	sig, err := s.kCmdSet.Sign(data)
	if err != nil {
		logger.Error("sign failed", "error", err)
		return err
	}

	s.write(fmt.Sprintf("SIGNATURE R: %x\n", sig.R()))
	s.write(fmt.Sprintf("SIGNATURE S: %x\n", sig.S()))
	s.write(fmt.Sprintf("SIGNATURE V: %x\n", sig.V()))

	return nil
}

func (s *Shell) commandKeycardSetPinlessPath(args ...string) error {
	if err := s.requireArgs(args, 1); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("set pinless path %s", args[0]))
	if err := s.kCmdSet.SetPinlessPath(args[0]); err != nil {
		logger.Error("set pinless path failed", "error", err)
		return err
	}

	return nil
}

func (s *Shell) requireArgs(args []string, possibleArgsN ...int) error {
	for _, n := range possibleArgsN {
		if len(args) == n {
			return nil
		}
	}

	ns := make([]string, len(possibleArgsN))
	for i, n := range possibleArgsN {
		ns[i] = fmt.Sprintf("%d", n)
	}

	return fmt.Errorf("wrong number of argument. got: %d, expected: %v", len(args), strings.Join(ns, " | "))
}

func (s *Shell) evalLine(rawLine string) error {
	line := strings.TrimSpace(rawLine)

	if len(line) == 0 || strings.HasPrefix(line, "#") {
		return nil
	}

	reg := regexp.MustCompile("\\s+")
	parts := reg.Split(line, -1)
	if cmd, ok := s.commands[parts[0]]; ok {
		return cmd(parts[1:]...)
	}

	return fmt.Errorf("command not found: %s", parts[0])
}

func (s *Shell) parseHex(str string) ([]byte, error) {
	if str[:2] == "0x" {
		str = str[2:]
	}

	return hex.DecodeString(str)
}
