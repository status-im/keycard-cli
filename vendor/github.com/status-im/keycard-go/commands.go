package keycard

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/status-im/keycard-go/apdu"
	"github.com/status-im/keycard-go/derivationpath"
	"github.com/status-im/keycard-go/globalplatform"
)

const (
	InsInit                 = uint8(0xFE)
	InsOpenSecureChannel    = uint8(0x10)
	InsMutuallyAuthenticate = uint8(0x11)
	InsPair                 = uint8(0x12)
	InsGetStatus            = uint8(0xF2)
	InsGenerateKey          = uint8(0xD4)
	InsVerifyPIN            = uint8(0x20)
	InsDeriveKey            = uint8(0xD1)
	InsSign                 = uint8(0xC0)
	InsSetPinlessPath       = uint8(0xC1)

	P1PairingFirstStep     = uint8(0x00)
	P1PairingFinalStep     = uint8(0x01)
	P1GetStatusApplication = uint8(0x00)
	P1GetStatusKeyPath     = uint8(0x01)
	P1DeriveKeyFromMaster  = uint8(0x00)
	P1DeriveKeyFromParent  = uint8(0x01)
	P1DeriveKeyFromCurrent = uint8(0x10)
)

func NewCommandInit(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsInit,
		uint8(0x00),
		uint8(0x00),
		data,
	)
}

func NewCommandPairFirstStep(challenge []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsPair,
		P1PairingFirstStep,
		uint8(0x00),
		challenge,
	)
}

func NewCommandPairFinalStep(cryptogramHash []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsPair,
		P1PairingFinalStep,
		uint8(0x00),
		cryptogramHash,
	)
}

func NewCommandOpenSecureChannel(pairingIndex uint8, pubKey []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsOpenSecureChannel,
		pairingIndex,
		uint8(0x00),
		pubKey,
	)
}

func NewCommandMutuallyAuthenticate(data []byte) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsMutuallyAuthenticate,
		uint8(0x00),
		uint8(0x00),
		data,
	)
}

func NewCommandGetStatus(p1 uint8) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsGetStatus,
		p1,
		uint8(0x00),
		[]byte{},
	)
}

func NewCommandGenerateKey() *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsGenerateKey,
		uint8(0),
		uint8(0),
		[]byte{},
	)
}

func NewCommandVerifyPIN(pin string) *apdu.Command {
	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsVerifyPIN,
		uint8(0),
		uint8(0),
		[]byte(pin),
	)
}

func NewCommandDeriveKey(pathStr string) (*apdu.Command, error) {
	startingPoint, path, err := derivationpath.Decode(pathStr)
	if err != nil {
		return nil, err
	}

	var p1 uint8
	switch startingPoint {
	case derivationpath.StartingPointMaster:
		p1 = P1DeriveKeyFromMaster
	case derivationpath.StartingPointParent:
		p1 = P1DeriveKeyFromParent
	case derivationpath.StartingPointCurrent:
		p1 = P1DeriveKeyFromCurrent
	default:
		return nil, fmt.Errorf("invalid startingPoint %d", startingPoint)
	}

	data := new(bytes.Buffer)
	for _, segment := range path {
		if err := binary.Write(data, binary.BigEndian, segment); err != nil {
			return nil, err
		}
	}

	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsDeriveKey,
		p1,
		uint8(0),
		data.Bytes(),
	), nil
}

func NewCommandSetPinlessPath(pathStr string) (*apdu.Command, error) {
	startingPoint, path, err := derivationpath.Decode(pathStr)
	if err != nil {
		return nil, err
	}

	if startingPoint != derivationpath.StartingPointMaster {
		return nil, fmt.Errorf("pinless path must be set with an absolute path")
	}

	data := new(bytes.Buffer)
	for _, segment := range path {
		if err := binary.Write(data, binary.BigEndian, segment); err != nil {
			return nil, err
		}
	}

	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsSetPinlessPath,
		uint8(0),
		uint8(0),
		data.Bytes(),
	), nil
}

func NewCommandSign(data []byte) (*apdu.Command, error) {
	if len(data) != 32 {
		return nil, fmt.Errorf("data length must be 32, got %d", len(data))
	}

	return apdu.NewCommand(
		globalplatform.ClaGp,
		InsSign,
		uint8(0),
		uint8(0),
		data,
	), nil
}
