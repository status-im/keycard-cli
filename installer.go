package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/globalplatform"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/status-im/keycard-go/identifiers"
	"github.com/status-im/keycard-go/types"
)

var (
	ErrAppletAlreadyInstalled = errors.New("keycard applet already installed")
	ndefRecord                = hexutils.HexToBytes("0024d40f12616e64726f69642e636f6d3a706b67696d2e7374617475732e657468657265756d")
)

// Installer defines a struct with methods to install applets in a card.
type Installer struct {
	c types.Channel
}

// NewInstaller returns a new Installer that communicates to Transmitter t.
func NewInstaller(t globalplatform.Transmitter) *Installer {
	return &Installer{
		c: globalplatform.NewNormalChannel(t),
	}
}

// Install installs the applet from the specified capFile.
func (i *Installer) Install(capFile *os.File, overwriteApplet bool) error {
	logger.Info("installation started")
	startTime := time.Now()
	cmdSet := globalplatform.NewCommandSet(i.c)

	logger.Info("check if keycard is already installed")
	if err := i.checkAppletAlreadyInstalled(cmdSet, overwriteApplet); err != nil {
		logger.Error("check if keycard is already installed failed", "error", err)
		return err
	}

	logger.Info("select ISD")
	err := cmdSet.Select()
	if err != nil {
		logger.Error("select failed", "error", err)
		return err
	}

	logger.Info("opening secure channel")
	if err = cmdSet.OpenSecureChannel(); err != nil {
		logger.Error("open secure channel failed", "error", err)
		return err
	}

	logger.Info("delete old version (if present)")
	if err = cmdSet.DeleteKeycardInstancesAndPackage(); err != nil {
		logger.Error("delete keycard instances and package failed", "error", err)
		return err
	}

	logger.Info("loading package")
	callback := func(index, total int) {
		logger.Debug(fmt.Sprintf("loading %d/%d", index+1, total))
	}
	if err = cmdSet.LoadKeycardPackage(capFile, callback); err != nil {
		logger.Error("load failed", "error", err)
		return err
	}

	logger.Info("installing NDEF applet")
	if err = cmdSet.InstallNDEFApplet(ndefRecord); err != nil {
		logger.Error("installing NDEF applet failed", "error", err)
		return err
	}

	logger.Info("installing Keycard applet")
	if err = cmdSet.InstallKeycardApplet(); err != nil {
		logger.Error("installing Keycard applet failed", "error", err)
		return err
	}

	elapsed := time.Now().Sub(startTime)
	logger.Info(fmt.Sprintf("installation completed in %f seconds", elapsed.Seconds()))
	return err
}

// Delete deletes the applet from the card.
func (i *Installer) Delete() error {
	cmdSet := globalplatform.NewCommandSet(i.c)

	logger.Info("select ISD")
	err := cmdSet.Select()
	if err != nil {
		logger.Error("select failed", "error", err)
		return err
	}

	logger.Info("opening secure channel")
	if err = cmdSet.OpenSecureChannel(); err != nil {
		logger.Error("open secure channel failed", "error", err)
		return err
	}

	logger.Info("delete old version")
	if err = cmdSet.DeleteKeycardInstancesAndPackage(); err != nil {
		logger.Error("delete keycard instances and package failed", "error", err)
		return err
	}

	return nil
}

func (i *Installer) checkAppletAlreadyInstalled(cmdSet *globalplatform.CommandSet, overwriteApplet bool) error {
	keycardInstanceAID, err := identifiers.KeycardInstanceAID(identifiers.KeycardDefaultInstanceIndex)
	if err != nil {
		return err
	}

	err = cmdSet.SelectAID(keycardInstanceAID)
	switch e := err.(type) {
	case *apdu.ErrBadResponse:
		// keycard applet not found, so not installed yet.
		if e.Sw == globalplatform.SwFileNotFound {
			return nil
		}
		return err
	case nil: // selected successfully, so it's already installed
		if overwriteApplet {
			return nil
		}
		return ErrAppletAlreadyInstalled
	default:
		return err
	}
}
