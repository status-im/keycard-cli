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
	"strings"
)

// Message represents an NDEF Message, which is a collection of one or
// more NDEF Records.
//
// Most common types of NDEF Messages (URI, Media) only have a single
// record. However, others, like Smart Posters, have multiple ones.
type Message struct {
	Records []*Record
}

// NewMessage returns a new Message initialized with a single Record
// with the TNF, Type, ID and Payload values.
func NewMessage(tnf byte, rtype string, id string, payload RecordPayload) *Message {
	return &Message{
		Records: []*Record{NewRecord(tnf, rtype, id, payload)},
	}
}

// NewMessageFromRecords returns a new Message containing several NDEF
// Records. The MB and ME flags for the records are adjusted to
// create a valid message.
func NewMessageFromRecords(records ...*Record) *Message {
	n := len(records)
	if n == 0 {
		return &Message{}
	}

	last := n - 1

	for _, r := range records {
		r.SetMB(false)
		r.SetME(false)
	}

	records[0].SetMB(true)
	records[last].SetME(true)

	return &Message{
		Records: records,
	}
}

// NewTextMessage returns a new Message with a single Record
// of WellKnownType T[ext].
func NewTextMessage(textVal, language string) *Message {
	return &Message{
		[]*Record{NewTextRecord(textVal, language)},
	}
}

// NewURIMessage returns a new Message with a single Record
// of WellKnownType U[RI].
func NewURIMessage(uriVal string) *Message {
	return &Message{
		[]*Record{NewURIRecord(uriVal)},
	}
}

// NewSmartPosterMessage returns a new Message with a single Record
// of WellKnownType Sp (Smart Poster).
func NewSmartPosterMessage(msgPayload *Message) *Message {
	return &Message{
		[]*Record{NewSmartPosterRecord(msgPayload)},
	}
}

// NewMediaMessage returns a new Message with a single Record
// of Media (RFC-2046) type.
//
// mimeType is something like "text/json" or "image/jpeg".
func NewMediaMessage(mimeType string, payload []byte) *Message {
	return &Message{
		[]*Record{NewMediaRecord(mimeType, payload)},
	}
}

// NewAbsoluteURIMessage returns a new Message with a single Record
// of AbsoluteURI type.
//
// AbsoluteURI means that the type of the payload for this record is
// defined by an URI resource. It is not supposed to be used to
// describe an URI. For that, use NewURIRecord().
func NewAbsoluteURIMessage(typeURI string, payload []byte) *Message {
	return &Message{
		[]*Record{NewAbsoluteURIRecord(typeURI, payload)},
	}
}

// NewExternalMessage returns a new Message with a single Record
// of NFC Forum External type.
func NewExternalMessage(extType string, payload []byte) *Message {
	return &Message{
		[]*Record{NewExternalRecord(extType, payload)},
	}
}

// Returns the string representation of each of the records in the message.
func (m *Message) String() string {
	str := ""
	last := len(m.Records) - 1
	for i, r := range m.Records {
		str += r.String()
		if i != last {
			str += "\n"
		}
	}
	return str
}

// Inspect returns a string with information about the message and its records.
func (m *Message) Inspect() string {
	str := fmt.Sprintf("NDEF Message with %d records.", len(m.Records))
	if len(m.Records) > 0 {
		str += "\n"
		for i, r := range m.Records {
			str += fmt.Sprintf("Record %d:\n", i)
			rIns := r.Inspect()
			rInsLines := strings.Split(rIns, "\n")
			for _, l := range rInsLines {
				str += "  " + l + "\n"
			}
		}
	}
	return str
}

// Unmarshal parses a byte slice into a Message. This is done by
// parsing all Records in the slice, until there are no more to parse.
//
// Returns the number of bytes processed (message length), or an error
// if something looks wrong with the message or its records.
func (m *Message) Unmarshal(buf []byte) (rLen int, err error) {
	m.Records = []*Record{}
	rLen = 0
	for rLen < len(buf) {
		r := new(Record)
		recordLen, err := r.Unmarshal(buf[rLen:])
		rLen += recordLen
		if err != nil {
			return rLen, err
		}
		m.Records = append(m.Records, r)
		if r.ME() { // last record in message
			break
		}
	}

	err = m.check()
	return rLen, err
}

// Marshal provides the byte slice representation of a Message,
// which is the concatenation of the Marshaling of each of its records.
//
// Returns an error if something goes wrong.
func (m *Message) Marshal() ([]byte, error) {
	err := m.check()
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	for _, r := range m.Records {
		rBytes, err := r.Marshal()
		if err != nil {
			return nil, err
		}
		_, err = buffer.Write(rBytes)
		if err != nil {
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

func (m *Message) check() error {
	last := len(m.Records) - 1

	if last < 0 {
		return errors.New(eNORECORDS)
	}

	if !m.Records[0].MB() {
		return errors.New(eNOMB)
	}

	if !m.Records[last].ME() {
		return errors.New(eNOME)
	}

	for i, r := range m.Records {
		if i > 0 && r.MB() {
			return errors.New(eBADMB)
		}
		if i < last && r.ME() {
			return errors.New(eBADME)
		}
	}

	return nil
}

// Check errors
const (
	eNORECORDS = "NDEF Message Check: No records"
	eNOMB      = "NDEF Message Check: first record has not the MessageBegin flag set"
	eNOME      = "NDEF Message Check: last record has not the MessageEnd flag set"
	eBADMB     = "NDEF Message Check: middle record has the MessageBegin flag set"
	eBADME     = "NDEF Message Check: middle record has the MessageEnd flag set"
)
