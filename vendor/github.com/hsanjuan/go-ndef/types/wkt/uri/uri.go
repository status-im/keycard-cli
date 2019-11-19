/***
    Copyright (c) 2018, Hector Sanjuan

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Lesser General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Lesser General Public License for more details.

    You should have received a copy of the GNU Lesser General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
***/

// Package uri provides support for NDEF Payloads of URI type.
// It follows the NFC Forum URI Record Type Definition specification
// (NFCForum-TS-RTD_URI_1.0).
//
// The Payload type implements the RecordPayload interface from ndef,
// so it can be used as ndef.Record.Payload.
package uri

import (
	"regexp"
	"strings"
)

// URIProtocols provides a mapping between the first byte if a NDEF Payload of
// type "U" (URI) and the string value for the protocol.
var URIProtocols = map[byte]string{
	0:  "",
	1:  "http://www.",
	2:  "https://www.",
	3:  "http://",
	4:  "https://",
	5:  "tel:",
	6:  "mailto:",
	7:  "ftp://anonymous:anonymous@",
	8:  "ftp://ftp.",
	9:  "ftps://",
	10: "sftp://",
	11: "smb://",
	12: "nfs://",
	13: "ftp://",
	14: "dev://",
	15: "news:",
	16: "telnet://",
	17: "imap:",
	18: "rtsp://",
	19: "urn:",
	20: "pop:",
	21: "sip:",
	22: "sips:",
	23: "tftp:",
	24: "btspp://",
	25: "btl2cap://",
	26: "btgoep://",
	27: "tcpobex://",
	28: "irdaobex://",
	29: "file://",
	30: "urn:epc:id:",
	31: "urn:epc:tag:",
	32: "urn:epc:pat:",
	33: "urn:epc:raw:",
	34: "urn:epc:",
	35: "urn:nfc:",
}

// Payload represents a NDEF Record Payload of Type "U".
type Payload struct {
	IdentCode byte
	URIField  string
}

// New returns a pointer to an Payload object.
// The Identifier code is automatically
// set based on the provided string.
func New(uriStr string) *Payload {
	u := new(Payload)
	u.URIField = uriStr
	for i := byte(1); i < 36; i++ {
		m, _ := regexp.MatchString("^"+URIProtocols[i], uriStr)
		if m {
			u.IdentCode = i
			u.URIField = strings.Replace(uriStr,
				URIProtocols[i], "", 1)
			break
		}
	}
	return u
}

// String returns the URI string.
func (u *Payload) String() string {
	return URIProtocols[u.IdentCode] + u.URIField
}

// Type returns the Uniform Resource Name for URIs.
func (u *Payload) Type() string {
	return "urn:nfc:wkt:U"
}

// Marshal returns the bytes representing the payload of a Record of URI type.
func (u *Payload) Marshal() []byte {
	p := []byte{u.IdentCode}
	return append(p, []byte(u.URIField)...)
}

// Unmarshal parses the payload of a URI type record.
func (u *Payload) Unmarshal(buf []byte) {
	u.IdentCode = 0
	u.URIField = ""
	if len(buf) > 0 {
		u.IdentCode = buf[0]
	}
	if len(buf) > 1 {
		u.URIField = string(buf[1:])
	}
}

// Len is the length of the byte slice resulting of Marshaling.
func (u *Payload) Len() int {
	return len(u.Marshal())
}
