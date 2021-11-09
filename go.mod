module github.com/status-im/keycard-cli

go 1.17

replace github.com/ethereum/go-ethereum v1.10.4 => github.com/status-im/go-ethereum v1.10.4-status.2

require (
	github.com/ebfe/scard v0.0.0-20190212122703-c3d1b1916a95
	github.com/ethereum/go-ethereum v1.10.12
	github.com/hsanjuan/go-ndef v0.0.1
	github.com/status-im/keycard-go v0.0.0-20211109104530-b0e0482ba91d
)

require (
	github.com/btcsuite/btcd v0.22.0-beta // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	golang.org/x/crypto v0.0.0-20211108221036-ceb1ce70b4fa // indirect
	golang.org/x/sys v0.0.0-20211109065445-02f5c0300f6e // indirect
	golang.org/x/text v0.3.7 // indirect
)
