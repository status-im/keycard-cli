package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"

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
	secrets, err := keycard.NewSecrets()
	if err != nil {
		return nil, err
	}

	info, err := keycard.Select(i.c, identifiers.KeycardAID)
	if err != nil {
		return nil, err
	}

	if !info.Installed {
		return nil, errAppletNotInstalled
	}

	if info.Initialized {
		return nil, errCardAlreadyInitialized
	}

	err = keycard.Init(i.c, info.PublicKey, secrets, identifiers.KeycardAID)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

// Info returns a types.ApplicationInfo struct with info about the card.
func (i *Initializer) Info() (types.ApplicationInfo, error) {
	cmdSet := keycard.NewCommandSet(i.c)
	err := cmdSet.Select()

	return cmdSet.ApplicationInfo, err
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

func (i *Initializer) deleteAID(aids ...[]byte) error {
	for _, aid := range aids {
		del := globalplatform.NewCommandDelete(aid)
		_, err := i.send("delete", del, globalplatform.SwOK, globalplatform.SwReferencedDataNotFound)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Initializer) installApplets(capFile *os.File) error {
	// install for load
	preLoad := globalplatform.NewCommandInstallForLoad(identifiers.PackageAID, identifiers.CardManagerAID)
	_, err := i.send("install for load", preLoad)
	if err != nil {
		return err
	}

	// load
	load, err := globalplatform.NewLoadCommandStream(capFile)
	if err != nil {
		return err
	}

	for load.Next() {
		cmd := load.GetCommand()
		_, err = i.send(fmt.Sprintf("load %d of 40", load.Index()+1), cmd)
		if err != nil {
			return err
		}
	}

	installNdef := globalplatform.NewCommandInstallForInstall(identifiers.PackageAID, identifiers.NdefAID, identifiers.NdefInstanceAID, []byte{})
	_, err = i.send("install for install (ndef)", installNdef)
	if err != nil {
		return err
	}

	instanceAID, err := identifiers.KeycardInstanceAID(1)
	if err != nil {
		return err
	}

	installWallet := globalplatform.NewCommandInstallForInstall(identifiers.PackageAID, identifiers.KeycardAID, instanceAID, []byte{})
	_, err = i.send("install for install (wallet)", installWallet)

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
