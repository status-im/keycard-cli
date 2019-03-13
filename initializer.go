package main

import (
	"crypto/rand"
	"errors"
	"fmt"

	keycard "github.com/status-im/keycard-go"
	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/globalplatform"
	"github.com/status-im/keycard-go/identifiers"
	"github.com/status-im/keycard-go/types"
)

var (
	errAppletNotInstalled     = errors.New("applet not installed")
	errCardNotInitialized     = errors.New("card not initialized")
	errCardAlreadyInitialized = errors.New("card already initialized")

	ErrNotInitialized = errors.New("card not initialized")
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

	secrets, err := keycard.NewSecrets()
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
	if e, ok := err.(*apdu.ErrBadResponse); ok && e.Sw == globalplatform.SwFileNotFound {
		err = nil
	} else {
		logger.Error("select failed", "error", err)
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
	cmdSet.SetPairingInfo(key, index)
	appStatus, err := cmdSet.GetStatus()
	if err != nil {
		logger.Error("get status failed", "error", err)
		return nil, err
	}

	return appStatus, nil
}

func (i *Initializer) initGPSecureChannel(sdaid []byte) error {
	// select card manager
	err := i.selectAID(sdaid)
	if err != nil {
		return err
	}

	// initialize update
	session, err := i.initializeUpdate()
	if err != nil {
		return err
	}

	i.c = globalplatform.NewSecureChannel(session, i.c)

	// external authenticate
	return i.externalAuthenticate(session)
}

func (i *Initializer) selectAID(aid []byte) error {
	sel := globalplatform.NewCommandSelect(identifiers.CardManagerAID)
	_, err := i.send("select", sel)

	return err
}

func (i *Initializer) initializeUpdate() (*globalplatform.Session, error) {
	hostChallenge, err := generateHostChallenge()
	if err != nil {
		return nil, err
	}

	init := globalplatform.NewCommandInitializeUpdate(hostChallenge)
	resp, err := i.send("initialize update", init)
	if err != nil {
		return nil, err
	}

	// verify cryptogram and initialize session keys
	keys := globalplatform.NewSCP02Keys(identifiers.CardTestKey, identifiers.CardTestKey)
	session, err := globalplatform.NewSession(keys, resp, hostChallenge)

	return session, err
}

func (i *Initializer) externalAuthenticate(session *globalplatform.Session) error {
	encKey := session.Keys().Enc()
	extAuth, err := globalplatform.NewCommandExternalAuthenticate(encKey, session.CardChallenge(), session.HostChallenge())
	if err != nil {
		return err
	}

	_, err = i.send("external authenticate", extAuth)

	return err
}

func (i *Initializer) send(description string, cmd *apdu.Command, allowedResponses ...uint16) (*apdu.Response, error) {
	logger.Debug("sending apdu command", "name", description)
	resp, err := i.c.Send(cmd)
	if err != nil {
		return nil, err
	}

	if len(allowedResponses) == 0 {
		allowedResponses = []uint16{apdu.SwOK}
	}

	for _, code := range allowedResponses {
		if code == resp.Sw {
			return resp, nil
		}
	}

	err = fmt.Errorf("unexpected response from command %s: %x", description, resp.Sw)

	return nil, err
}

func generateHostChallenge() ([]byte, error) {
	c := make([]byte, 8)
	_, err := rand.Read(c)
	return c, err
}
