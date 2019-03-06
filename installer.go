package main

import (
	"fmt"
	"os"
	"time"

	"github.com/status-im/keycard-go/globalplatform"
	"github.com/status-im/keycard-go/hexutils"
)

// Installer defines a struct with methods to install applets in a card.
type Installer struct {
	c globalplatform.Channel
}

// NewInstaller returns a new Installer that communicates to Transmitter t.
func NewInstaller(t globalplatform.Transmitter) *Installer {
	return &Installer{
		c: globalplatform.NewNormalChannel(t),
	}
}

// Install installs the applet from the specified capFile.
func (i *Installer) Install(capFile *os.File, overwriteApplet bool) (err error) {
	logger.Debug("installation started")
	startTime := time.Now()
	cmdSet := globalplatform.NewCommandSet(i.c)

	logger.Debug("select ISD")
	if err = cmdSet.Select(); err != nil {
		logger.Error("select failed", "error", err)
		return err
	}

	logger.Debug("opening secure channel")
	if err = cmdSet.OpenSecureChannel(); err != nil {
		logger.Error("open secure channel failed", "error", err)
		return err
	}

	logger.Debug("delete old version (if present)")
	if err = cmdSet.DeleteKeycardInstancesAndPackage(); err != nil {
		logger.Error("delete keycard instances and package failed", "error", err)
		return err
	}

	logger.Debug("loading package")
	callback := func(index, total int) {
		logger.Debug(fmt.Sprintf("loading %d/%d", index+1, total))
	}
	if err = cmdSet.LoadKeycardPackage(capFile, callback); err != nil {
		logger.Error("load failed", "error", err)
		return err
	}

	logger.Debug("installing NDEF applet")
	ndefRecord := hexutils.HexToBytes("0024d40f12616e64726f69642e636f6d3a706b67696d2e7374617475732e657468657265756d")
	if err = cmdSet.InstallNDEFApplet(ndefRecord); err != nil {
		logger.Error("installing NDEF applet failed", "error", err)
		return err
	}

	logger.Debug("installing Keycard applet")
	if err = cmdSet.InstallKeycardApplet(); err != nil {
		logger.Error("installing Keycard applet failed", "error", err)
		return err
	}

	elapsed := time.Now().Sub(startTime)
	logger.Info(fmt.Sprintf("installation completed in %f seconds", elapsed.Seconds()))
	return err
}
