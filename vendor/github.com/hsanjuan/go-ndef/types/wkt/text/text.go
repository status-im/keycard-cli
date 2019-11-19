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

// BUG(hector): The implementation ignores the guidelines about displaying the
// text and removing the control characters.

// BUG(hector): UTF-16 with different byte order, with/without BOM is not tested.

// Package text provides support for NDEF Payloads of Text type.
// It follows the NFC Forum Text Record Type Definition specification
// (NFCForum-TS-RTD_Text_1.0).
//
// The Payload type implements the RecordPayload interface from ndef,
// so it can be used as ndef.Record.Payload.
package text

import (
	"bytes"
	"strings"
	"unicode/utf16"
)

// Payload represents a NDEF Record Payload of type "T", which
// holds a text field and IANA-formatted language information.
type Payload struct {
	Language string
	Text     string
}

// New returns a pointer to a Payload.
//
// The language parameter must be compliant to RFC 3066 (i.e. "en_US"),
// but no check is performed.
func New(text, language string) *Payload {
	return &Payload{
		Language: language,
		Text:     text,
	}
}

// String returns the actual text. Language information is ommited.
func (t *Payload) String() string {
	return t.Text
}

// Type returns the URN for Text types.
func (t *Payload) Type() string {
	return "urn:nfc:wkt:T"
}

// Marshal returns the bytes representing the payload of a text Record.
func (t *Payload) Marshal() []byte {
	var buf bytes.Buffer
	ianaLen := byte(len(t.Language))
	buf.WriteByte(ianaLen)
	buf.Write([]byte(t.Language))
	buf.Write([]byte(t.Text))
	return buf.Bytes()
}

// Unmarshal parses the Payload from a text Record.
func (t *Payload) Unmarshal(buf []byte) {
	t.Language = ""
	t.Text = ""
	i := byte(0)
	if len(buf) < 1 {
		return
	}
	firstByte := buf[i]
	i++
	isUtf16 := firstByte>>7 == 1
	// firstByte>>6 must be set to 0
	ianaLen := 0x3F & firstByte // last 5 bytes
	if len(buf) < int(i+ianaLen) {
		return
	}
	t.Language = string(buf[i : i+ianaLen])
	i += ianaLen
	if len(buf) < int(i) {
		return
	}
	if isUtf16 {
		//Convert buf to []uint16
		bytesBuf := bytes.NewBuffer(buf[i:])
		var finished bool
		var uint16buf []uint16
		for !finished {
			b1, err := bytesBuf.ReadByte()
			b2, err := bytesBuf.ReadByte()
			uint16buf = append(uint16buf,
				uint16(b1<<8)|uint16(b2))
			finished = err != nil
		}
		runes := utf16.Decode(uint16buf)
		// It appears we get an extra trailing char at the end
		t.Text = strings.TrimSuffix(string(runes), "\x00")
	} else {
		t.Text = string(buf[i:])
	}
}

// Len is the length of the byte slice resulting of Marshaling..
func (t *Payload) Len() int {
	return len(t.Marshal())
}
