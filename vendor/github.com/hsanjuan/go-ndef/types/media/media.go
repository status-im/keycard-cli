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

// Package media provides an implementation for NDEF Payloads for media
// types.
package media

// Payload is a wrapper to store a Payload
type Payload struct {
	MimeType string
	Payload  []byte
}

// New returns a pointer to a Payload type holding the given payload with the
// given type.
func New(mimeType string, payload []byte) *Payload {
	return &Payload{
		MimeType: mimeType,
		Payload:  payload,
	}
}

// String returns a string explaining that we are not sure how to print
// this type.
func (media *Payload) String() string {
	if media.Len() > 0 {
		return "<The message contains a payload>"
	}
	return ""
}

// Type returns the mime type of this payload.
func (media *Payload) Type() string {
	return media.MimeType
}

// Marshal returns the bytes representing the payload
func (media *Payload) Marshal() []byte {
	return media.Payload
}

// Unmarshal parses a generic payload
func (media *Payload) Unmarshal(buf []byte) {
	media.Payload = buf
}

// Len is the length of the byte slice resulting of Marshaling
// this Payload.
func (media *Payload) Len() int {
	return len(media.Marshal())
}
