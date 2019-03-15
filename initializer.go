package main

import (
	"errors"

	keycard "github.com/status-im/keycard-go"
	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/globalplatform"
	"github.com/status-im/keycard-go/types"
)

var (
	errAppletNotInstalled     = errors.New("applet not installed")
	errCardNotInitialized     = errors.New("card not initialized")
	errCardAlreadyInitialized = errors.New("card already initialized")

	ErrNotInitialized = errors.New("card not initialized")
	ErrNotInstalled   = errors.New("applet not initialized")
)

// Initializer defines a struct with methods to install applets and initialize a card.
type Initializer struct {
	c types.Channel
}

// NewInitializer returns a new Initializer that communicates to Transmitter t.
func NewInitializer(t globalplatform.Transmitter) *Initializer {
	return &Initializer{
		c: globalplatform.NewNormalChannel(t),
	}
}

func (i *Initializer) Init() (*keycard.Secrets, error) {
	logger.Info("initialization started")
	cmdSet := keycard.NewCommandSet(i.c)

	secrets, err := keycard.GenerateSecrets()
	if err != nil {
		return nil, err
	}

	logger.Info("select keycard applet")
	err = cmdSet.Select()
	if err != nil {
		logger.Error("select failed", "error", err)
		return nil, err
	}

	if !cmdSet.ApplicationInfo.Installed {
		logger.Error("initialization failed", "error", errAppletNotInstalled)
		return nil, errAppletNotInstalled
	}

	if cmdSet.ApplicationInfo.Initialized {
		logger.Error("initialization failed", "error", errCardAlreadyInitialized)
		return nil, errCardAlreadyInitialized
	}

	logger.Info("initializing")
	err = cmdSet.Init(secrets)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

// Info returns a types.ApplicationInfo struct with info about the card.
func (i *Initializer) Info() (*types.ApplicationInfo, error) {
	logger.Info("info started")
	cmdSet := keycard.NewCommandSet(i.c)

	logger.Info("select keycard applet")
	err := cmdSet.Select()
	if err != nil {
		if e, ok := err.(*apdu.ErrBadResponse); ok && e.Sw == globalplatform.SwFileNotFound {
			err = nil
		} else {
			logger.Error("select failed", "error", err)
		}
	}

	return cmdSet.ApplicationInfo, err
}

func (i *Initializer) Pair(pairingPass string) (*types.PairingInfo, error) {
	logger.Info("pairing started")
	cmdSet := keycard.NewCommandSet(i.c)

	logger.Info("select keycard applet")
	err := cmdSet.Select()
	if err != nil {
		logger.Error("select failed", "error", err)
		return nil, err
	}

	if !cmdSet.ApplicationInfo.Initialized {
		logger.Error("pairing failed", "error", ErrNotInitialized)
		return nil, ErrNotInitialized
	}

	logger.Info("pairing")
	err = cmdSet.Pair(pairingPass)
	return cmdSet.PairingInfo, err
}

func (i *Initializer) Status(key []byte, index int) (*types.ApplicationStatus, error) {
	logger.Info("pairing started")
	cmdSet := keycard.NewCommandSet(i.c)

	logger.Info("select keycard applet")
	err := cmdSet.Select()
	if err != nil {
		logger.Error("select failed", "error", err)
		return nil, err
	}

	if !cmdSet.ApplicationInfo.Initialized {
		logger.Error("pairing failed", "error", ErrNotInitialized)
		return nil, ErrNotInitialized
	}

	logger.Info("open secure channel")
	cmdSet.SetPairingInfo(key, index)
	err = cmdSet.OpenSecureChannel()
	if err != nil {
		logger.Error("open secure channel failed", "error", err)
		return nil, err
	}

	logger.Info("get status")
	appStatus, err := cmdSet.GetStatusApplication()
	if err != nil {
		logger.Error("get status failed", "error", err)
		return nil, err
	}

	return appStatus, nil
}
