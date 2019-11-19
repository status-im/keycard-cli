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

// Package absoluteuri provides an implementation for NDEF Payloads
// of Absolute URI type.
package absoluteuri

// Payload is a wrapper to store a Payload
type Payload struct {
	URIType string
	Payload []byte
}

// New returns a pointer to a Payload type holding the given payload with the
// given type.
func New(uriType string, payload []byte) *Payload {
	return &Payload{
		URIType: uriType,
		Payload: payload,
	}
}

// String returns a string explaining that we are not sure how to print
// this type.
func (absu *Payload) String() string {
	if absu.Len() > 0 {
		return "<The message contains a binary payload>"
	}
	return ""
}

// Type returns the URI defining the type for this payload
func (absu *Payload) Type() string {
	return absu.URIType
}

// Marshal returns the bytes representing the payload
func (absu *Payload) Marshal() []byte {
	return absu.Payload
}

// Unmarshal parses a generic payload
func (absu *Payload) Unmarshal(buf []byte) {
	absu.Payload = buf
}

// Len is the length of the byte slice resulting of Marshaling
// this Payload.
func (absu *Payload) Len() int {
	return len(absu.Marshal())
}
