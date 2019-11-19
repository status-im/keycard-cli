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

package ndef

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
)

// RecordChunk represents how a Record is actually stored
// We need this for parsing and checking the validity of a Record
// before assembling them.
type recordChunk struct {
	// First byte
	MB            bool   // Message begin
	ME            bool   // Message end
	CF            bool   // Chunk Flag
	SR            bool   // Short record
	IL            bool   // ID length field present
	TNF           byte   // Type name format (3 bits)
	TypeLength    byte   // Type Length
	IDLength      byte   // Length of the ID field
	PayloadLength uint64 // Length of the Payload.
	Type          string // Type of the payload. Must follow TNF
	ID            string // An URI (per RFC 3986)
	Payload       []byte // Payload
}

// Reset clears up all the fields of the Record and sets them to their
// default values.
func (r *recordChunk) Reset() {
	r.MB = false
	r.ME = false
	r.CF = false
	r.SR = false
	r.IL = false
	r.TNF = 0
	r.TypeLength = 0
	r.IDLength = 0
	r.PayloadLength = 0
	r.Type = ""
	r.ID = ""
	r.Payload = []byte{}
}

// New returns a single RecordChunk with the given options. This chunk can be
// used directly to make a single-chunk NDEF Record on a single-record NDEF
// message: the MB and ME fields are set to true.
func newChunk(tnf byte, typ string, id string, payload []byte) *recordChunk {
	chunk := &recordChunk{}
	chunk.Reset()
	chunk.MB = true        // Message-begin
	chunk.ME = true        // Message-end
	chunk.CF = false       // not chunked
	chunk.IL = len(id) > 0 // only if ID field present
	chunk.TNF = tnf
	chunk.TypeLength = byte(len([]byte(typ)))
	chunk.Type = typ
	chunk.IDLength = byte(len([]byte(id)))
	chunk.ID = id

	payloadLen := uint64(len(payload))
	if payloadLen > 4294967295 { //2^32-1. 4GB message max.
		payloadLen = 4294967295
	}
	chunk.SR = payloadLen < 256 // Short record vs. Long
	chunk.PayloadLength = payloadLen

	// FIXME: If payload is greater than 2^32 - 1
	// we'll truncate without warning.
	chunk.Payload = payload[:payloadLen]
	return chunk
}

// Provide a string with information about this record chunk.
// Records' payload do not make sense without having compiled a whole Record
// so they are not dealed with here.
func (r *recordChunk) String() string {
	var str string
	str += fmt.Sprintf("MB: %t | ME: %t | CF: %t | SR: %t | IL: %t | TNF: %d\n",
		r.MB, r.ME, r.CF, r.SR, r.IL, r.TNF)
	str += fmt.Sprintf("TypeLength: %d", r.TypeLength)
	str += fmt.Sprintf(" | Type: %s\n", r.Type)
	str += fmt.Sprintf("Record Payload Length: %d",
		r.PayloadLength)
	if r.IL {
		str += fmt.Sprintf(" | IDLength: %d", r.IDLength)
		str += fmt.Sprintf(" | ID: %s", r.ID)
	}
	str += fmt.Sprintf("\n")
	return str
}

// Unmarshal parses a byte slice into a single Record chunk struct (the slice
// can have extra bytes which are ignored). The Record is always reset before
// parsing.
//
// Returns how many bytes were parsed from the slice (record length) or
// an error if something went wrong.
func (r *recordChunk) Unmarshal(buf []byte) (rLen int, err error) {
	// Handle errors that are produced by getByte() and getBytes()
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
			err = errors.New("Record.Unmarshal: " + err.Error())
		}
	}()
	r.Reset()
	bytesBuf := bytes.NewBuffer(buf)

	firstByte := getByte(bytesBuf)
	r.MB = (firstByte >> 7 & 0x1) == 1
	r.ME = (firstByte >> 6 & 0x1) == 1
	r.CF = (firstByte >> 5 & 0x1) == 1
	r.SR = (firstByte >> 4 & 0x1) == 1
	r.IL = (firstByte >> 3 & 0x1) == 1
	r.TNF = firstByte & 0x7

	r.TypeLength = getByte(bytesBuf)

	if r.SR { //This is a short record
		r.PayloadLength = uint64(getByte(bytesBuf))
	} else { // Regular record
		r.PayloadLength = bytesToUint64(getBytes(bytesBuf, 4))
	}
	if r.IL {
		r.IDLength = getByte(bytesBuf)
	}
	r.Type = string(getBytes(bytesBuf, int(r.TypeLength)))
	if r.IL {
		r.ID = string(getBytes(bytesBuf, int(r.IDLength)))
	}
	r.Payload = getBytes(bytesBuf, int(r.PayloadLength))

	rLen = len(buf) - bytesBuf.Len()
	err = r.Check()
	if err != nil {
		return rLen, err
	}
	return rLen, nil
}

// Marshal returns the byte representation of a Record, or an error
// if something went wrong
func (r *recordChunk) Marshal() ([]byte, error) {
	err := r.Check()
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	firstByte := byte(0)
	if r.MB {
		firstByte |= 0x1 << 7
	}
	if r.ME {
		firstByte |= 0x1 << 6
	}
	if r.CF {
		firstByte |= 0x1 << 5
	}
	if r.SR {
		firstByte |= 0x1 << 4
	}
	if r.IL {
		firstByte |= 0x1 << 3
	}
	firstByte |= (r.TNF & 0x7) //Last 3 bits are from TNF
	buffer.WriteByte(firstByte)
	// TypeLength byte
	buffer.WriteByte(r.TypeLength)

	// Payload Length byte (for SR) or 4 bytes for the regular case
	if r.SR {
		buffer.WriteByte(byte(r.PayloadLength))
	} else {
		buffer.Write(uint64ToBytes(r.PayloadLength, 4))
	}

	// ID Length byte if we are meant to have it
	if r.IL {
		buffer.WriteByte(r.IDLength)
	}

	// Write the type bytes if we have something
	if r.TypeLength > 0 {
		buffer.Write([]byte(r.Type))
	}

	// Write the ID bytes if we have something
	if r.IL && r.IDLength > 0 {
		buffer.Write([]byte(r.ID))
	}

	buffer.Write(r.Payload)
	return buffer.Bytes(), nil
}

// Check verifies that fields in this chunk are not in violation of the spec.
func (r *recordChunk) Check() error {
	// If the TNF value is 0x00, the TYPE_LENGTH, ID_LENGTH,
	// and PAYLOAD_LENGTH fields MUST be zero and the TYPE, ID,
	// and PAYLOAD fields MUST be omitted from the record.
	if r.TNF == Empty && (r.TypeLength > 0 ||
		r.IDLength > 0 || r.PayloadLength > 0) {
		return errors.New("Record.check: " +
			"Empty record TNF but not empty fields")
	}
	// If the TNF value is 0x05 or 0x06 (Unknown/Unchanged),
	// the TYPE_LENGTH field MUST be 0 and the TYPE
	// field MUST be omitted from the NDEF record.
	if (r.TNF == Unknown || r.TNF == Unchanged) && r.TypeLength > 0 {
		return errors.New("Record.check: " +
			"This TNF does not support a Type field")
	}

	// The TNF value MUST NOT be 0x07.
	if r.TNF == Reserved {
		return errors.New("Record.check: " +
			"The TNF cannot be Reserved, that value is reserved.")
	}

	// NFC Record Type Definition 3.4:
	// The binary encoding of Well Known Types
	// (including Global and Local Names) and External
	// Type names MUST be done according to the
	// ASCII chart in Appendix A.
	if r.TNF == NFCForumWellKnownType ||
		r.TNF == NFCForumExternalType {
		typeString := string(r.Type)
		for _, rune := range typeString {
			if rune < 32 || rune > 126 {
				return errors.New("Record.check(): " +
					"Record type names SHALL " +
					"be formed of characters from of the " +
					"US ASCII [ASCII] character set")
			}
		}
	}

	if r.IL && r.IDLength > 0 {
		for _, rune := range r.ID {
			if rune < 32 || rune > 126 {
				return errors.New("Record.check(): " +
					"ID must use ASCII characters")
			}
		}
	}
	return nil
}
