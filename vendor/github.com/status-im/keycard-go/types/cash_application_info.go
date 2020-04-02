package types

import "github.com/status-im/keycard-go/apdu"

type CashApplicationInfo struct {
	Installed  bool
	PublicKey  []byte
	PublicData []byte
	Version    []byte
}

func ParseCashApplicationInfo(data []byte) (*CashApplicationInfo, error) {
	info := &CashApplicationInfo{}

	if data[0] != TagApplicationInfoTemplate {
		return nil, ErrWrongApplicationInfoTemplate
	}

	info.Installed = true

	pubKey, err := apdu.FindTag(data, apdu.Tag{TagApplicationInfoTemplate}, apdu.Tag{0x80})
	if err != nil {
		return nil, err
	}

	pubData, err := apdu.FindTag(data, apdu.Tag{TagApplicationInfoTemplate}, apdu.Tag{0x82})
	if err != nil {
		return nil, err
	}

	appVersion, err := apdu.FindTag(data, apdu.Tag{TagApplicationInfoTemplate}, apdu.Tag{0x02})
	if err != nil {
		return nil, err
	}

	info.PublicKey = pubKey
	info.PublicData = pubData
	info.Version = appVersion

	return info, nil
}
