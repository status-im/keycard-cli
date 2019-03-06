package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"

	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/globalplatform"
	"github.com/status-im/keycard-go/lightwallet"
	"github.com/status-im/keycard-go/lightwallet/actions"
)

var (
	errAppletNotInstalled     = errors.New("applet not installed")
	errCardNotInitialized     = errors.New("card not initialized")
	errCardAlreadyInitialized = errors.New("card already initialized")

	ErrNotInitialized = errors.New("card not initialized")
)

// Initializer defines a struct with methods to install applets and initialize a card.
type Initializer struct {
	c globalplatform.Channel
}

// NewInitializer returns a new Initializer that communicates to Transmitter t.
func NewInitializer(t globalplatform.Transmitter) *Initializer {
	return &Initializer{
		c: globalplatform.NewNormalChannel(t),
	}
}

// Install installs the applet from the specified capFile.
func (i *Initializer) Install(capFile *os.File, overwriteApplet bool) error {
	info, err := actions.Select(i.c, lightwallet.WalletAID)
	if err != nil {
		return err
	}

	if info.Installed && !overwriteApplet {
		return errors.New("applet already installed")
	}

	err = i.initGPSecureChannel(lightwallet.CardManagerAID)
	if err != nil {
		return err
	}

	err = i.deleteAID(lightwallet.NdefInstanceAID, lightwallet.WalletInstanceAID, lightwallet.AppletPkgAID)
	if err != nil {
		return err
	}

	err = i.installApplets(capFile)
	if err != nil {
		return err
	}

	return err
}

func (i *Initializer) Init() (*lightwallet.Secrets, error) {
	secrets, err := lightwallet.NewSecrets()
	if err != nil {
		return nil, err
	}

	info, err := actions.Select(i.c, lightwallet.WalletAID)
	if err != nil {
		return nil, err
	}

	if !info.Installed {
		return nil, errAppletNotInstalled
	}

	if info.Initialized {
		return nil, errCardAlreadyInitialized
	}

	err = actions.Init(i.c, info.PublicKey, secrets, lightwallet.WalletAID)
	if err != nil {
		return nil, err
	}

	return secrets, nil
}

func (i *Initializer) Pair(pairingPass, pin string) (*lightwallet.PairingInfo, error) {
	appInfo, err := actions.Select(i.c, lightwallet.WalletAID)
	if err != nil {
		return nil, err
	}

	if !appInfo.Initialized {
		return nil, ErrNotInitialized
	}

	return actions.Pair(i.c, pairingPass, pin)
}

// Info returns a lightwallet.ApplicationInfo struct with info about the card.
func (i *Initializer) Info() (*lightwallet.ApplicationInfo, error) {
	return actions.Select(i.c, lightwallet.WalletAID)
}

// Status returns
func (i *Initializer) Status(index uint8, key []byte) (*lightwallet.ApplicationStatus, error) {
	info, err := actions.Select(i.c, lightwallet.WalletAID)
	if err != nil {
		return nil, err
	}

	if !info.Installed {
		return nil, errAppletNotInstalled
	}

	if !info.Initialized {
		return nil, errCardNotInitialized
	}

	sc, err := actions.OpenSecureChannel(i.c, info, index, key)
	if err != nil {
		return nil, err
	}

	return actions.GetStatusApplication(sc)
}

// Delete deletes the applet and related package from the card.
func (i *Initializer) Delete() error {
	err := i.initGPSecureChannel(lightwallet.CardManagerAID)
	if err != nil {
		return err
	}

	return i.deleteAID(lightwallet.NdefInstanceAID, lightwallet.WalletInstanceAID, lightwallet.AppletPkgAID)
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
	sel := globalplatform.NewCommandSelect(lightwallet.CardManagerAID)
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
	keys := globalplatform.NewSCP02Keys(lightwallet.CardTestKey, lightwallet.CardTestKey)
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
	preLoad := globalplatform.NewCommandInstallForLoad(lightwallet.AppletPkgAID, lightwallet.CardManagerAID)
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

	installNdef := globalplatform.NewCommandInstallForInstall(lightwallet.AppletPkgAID, lightwallet.NdefAppletAID, lightwallet.NdefInstanceAID, []byte{})
	_, err = i.send("install for install (ndef)", installNdef)
	if err != nil {
		return err
	}

	installWallet := globalplatform.NewCommandInstallForInstall(lightwallet.AppletPkgAID, lightwallet.WalletAID, lightwallet.WalletInstanceAID, []byte{})
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
