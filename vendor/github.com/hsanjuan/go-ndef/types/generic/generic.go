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

// Package generic provides a generic implementation for NDEF Payloads which are
// either custom or not supported yet.
package generic

// Payload is a wrapper to store a Payload.
type Payload struct {
	Payload []byte
}

// New returns a pointer to a Payload type holding the given payload.
func New(payload []byte) *Payload {
	return &Payload{
		Payload: payload,
	}
}

// String returns a string explaining that we are not sure how to print
// this type.
func (g *Payload) String() string {
	if g.Len() > 0 {
		return "<The message contains a binary payload>"
	}
	return ""
}

// Type returns the name of the payload type. In this case, it's generic.
func (g *Payload) Type() string {
	return "go-ndef-generic"
}

// Marshal returns the bytes representing the payload of a Record of
// generic type.
func (g *Payload) Marshal() []byte {
	return g.Payload
}

// Unmarshal parses a generic payload.
func (g *Payload) Unmarshal(buf []byte) {
	g.Payload = buf
}

// Len is the length of the byte slice resulting of Marshaling.
func (g *Payload) Len() int {
	return len(g.Marshal())
}
