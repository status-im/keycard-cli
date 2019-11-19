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

// Unfortunately splitting this to its own package causes a hard to break cycle.

// SmartPosterPayload represents the Payload of a Smart Poster, which is
// an NDEF Message with one or multiple records.
type SmartPosterPayload struct {
	Message *Message
}

// NewSmartPosterPayload returns a new Smart Poster payload.
func NewSmartPosterPayload(msg *Message) *SmartPosterPayload {
	return &SmartPosterPayload{
		Message: msg,
	}
}

// String returns the contents of the message contained in the Smart Poster.
func (sp *SmartPosterPayload) String() string {
	str := "\n"
	str += sp.Message.String()
	return str
}

// Type returns the URN for the Smart Poster type
func (sp *SmartPosterPayload) Type() string {
	return "urn:nfc:wkt:Sp"
}

// Marshal returns the bytes representing the payload of a Smart Poster.
// The payload is the contained NDEF Message.
func (sp *SmartPosterPayload) Marshal() []byte {
	bs, _ := sp.Message.Marshal()
	return bs
}

// Unmarshal parses the SmartPosterPayload from a Smart Poster.
func (sp *SmartPosterPayload) Unmarshal(buf []byte) {
	msg := &Message{}
	msg.Unmarshal(buf)
	sp.Message = msg
}

// Len returns the length of this payload in bytes.
func (sp *SmartPosterPayload) Len() int {
	return len(sp.Marshal())
}
