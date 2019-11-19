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
	"github.com/hsanjuan/go-ndef/types/absoluteuri"
	"github.com/hsanjuan/go-ndef/types/ext"
	"github.com/hsanjuan/go-ndef/types/generic"
	"github.com/hsanjuan/go-ndef/types/media"
	"github.com/hsanjuan/go-ndef/types/wkt/text"
	"github.com/hsanjuan/go-ndef/types/wkt/uri"
)

// The RecordPayload interface should be implemented by supported
// NDEF Record types. It ensures that we have a way to interpret payloads
// into printable information and to produce NDEF Record payloads for a given
// type.
type RecordPayload interface {
	// Returns a string representation of the Payload
	String() string
	// Provides serialization for the Payload
	Marshal() []byte
	// Provides de-serialization for the Payload
	Unmarshal(buf []byte)
	// Returns a string indetifying the type of this payload
	Type() string
	// Returns the length of the Payload (serialized)
	Len() int
}

func makeRecordPayload(tnf byte, rtype string, payload []byte) RecordPayload {
	var r RecordPayload
	switch tnf {
	case NFCForumWellKnownType:
		switch rtype {
		case "U":
			r = new(uri.Payload)
		case "T":
			r = new(text.Payload)
		case "Sp":
			r = new(SmartPosterPayload)
		default:
			r = new(generic.Payload)
		}
	case MediaType:
		r = media.New(rtype, nil)
	case NFCForumExternalType:
		r = ext.New(rtype, nil)
	case AbsoluteURI:
		r = absoluteuri.New(rtype, nil)
	default:
		r = new(generic.Payload)
	}
	r.Unmarshal(payload)
	return r
}
