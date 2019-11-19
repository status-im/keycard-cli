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

	"github.com/hsanjuan/go-ndef/types/absoluteuri"
	"github.com/hsanjuan/go-ndef/types/ext"
	"github.com/hsanjuan/go-ndef/types/media"
	"github.com/hsanjuan/go-ndef/types/wkt/text"
	"github.com/hsanjuan/go-ndef/types/wkt/uri"
)

// A Record is an NDEF Record. Multiple records can be
// part of a single NDEF Message.
type Record struct {
	chunks []*recordChunk
}

// NewRecord returns a single-chunked record with the given options.
// Use a generic.Payload if you want to use a custom byte-slice for payload.
func NewRecord(tnf byte, typ string, id string, payload RecordPayload) *Record {
	var payloadBytes []byte
	if payload != nil {
		payloadBytes = payload.Marshal()
	}

	chunk := newChunk(
		tnf,
		typ,
		id,
		payloadBytes,
	)
	return &Record{
		chunks: []*recordChunk{chunk},
	}
}

// TNF returns the Type Name Format (3 bits) associated to this Record.
func (r *Record) TNF() byte {
	if r.Empty() {
		return 0
	}
	return r.chunks[0].TNF
}

// Type returns the declared Type for this record.
func (r *Record) Type() string {
	if r.Empty() {
		return ""
	}
	return r.chunks[0].Type
}

// ID returns the declared record ID for this record.
func (r *Record) ID() string {
	if r.Empty() {
		return ""
	}
	return r.chunks[0].ID
}

// Payload returns the RecordPayload for this record. It will use
// one of the supported types, or otherwise a generic.Payload.
func (r *Record) Payload() (RecordPayload, error) {
	if r.Empty() {
		return nil, errors.New("empty record")
	}

	var buf bytes.Buffer
	for _, chunk := range r.chunks {
		_, err := buf.Write(chunk.Payload)
		if err != nil {
			return nil, err
		}
	}
	return makeRecordPayload(r.TNF(), r.Type(), buf.Bytes()), nil
}

// Empty returns true if this record has no chunks.
func (r *Record) Empty() bool {
	return len(r.chunks) == 0
}

// MB returns the value of the MessageBegin bit of the first chunk of this
// record, signaling that this is the first record in an NDEF Message.
// a NDEF Message.
func (r *Record) MB() bool {
	if r.Empty() {
		return false
	}
	return r.chunks[0].MB
}

// SetMB sets the MessageBegin bit of the first chunk of this Record.
func (r *Record) SetMB(b bool) {
	if r.Empty() {
		return
	}
	r.chunks[0].MB = b
}

// ME returns the value of the MessageEnd bit of the last chunk of this Record,
// signaling that this is the last record in an NDEF Message.
func (r *Record) ME() bool {
	if r.Empty() {
		return false
	}
	return r.chunks[len(r.chunks)-1].ME
}

// SetME sets the MessageEnd bit of the last chunk of this record.
func (r *Record) SetME(b bool) {
	if r.Empty() {
		return
	}
	r.chunks[len(r.chunks)-1].ME = b
}

// NewTextRecord returns a new Record with a
// Payload of Text [Well-Known] Type.
func NewTextRecord(textVal, language string) *Record {
	pl := text.New(textVal, language)
	return NewRecord(NFCForumWellKnownType, "T", "", pl)
}

// NewURIRecord returns a new Record with a
// Payload of URI [Well-Known] Type.
func NewURIRecord(uriVal string) *Record {
	pl := uri.New(uriVal)
	return NewRecord(NFCForumWellKnownType, "U", "", pl)
}

// NewSmartPosterRecord creates a new Record representing a Smart Poster.
// The Payload of a Smart Poster is an NDEF Message.
func NewSmartPosterRecord(msg *Message) *Record {
	pl := NewSmartPosterPayload(msg)
	return NewRecord(NFCForumWellKnownType, "Sp", "", pl)
}

// NewMediaRecord returns a new Record with a
// Media type (per RFC-2046) as payload.
//
// mimeType is something like "text/json" or "image/jpeg".
func NewMediaRecord(mimeType string, payload []byte) *Record {
	pl := media.New(mimeType, payload)
	return NewRecord(MediaType, mimeType, "", pl)
}

// NewAbsoluteURIRecord returns a new Record with a
// Payload of Absolute URI type.
//
// AbsoluteURI means that the type of the payload for this record is
// defined by an URI resource. It is not supposed to be used to
// describe an URI. For that, use NewURIRecord().
func NewAbsoluteURIRecord(typeURI string, payload []byte) *Record {
	pl := absoluteuri.New(typeURI, payload)
	return NewRecord(AbsoluteURI, typeURI, "", pl)
}

// NewExternalRecord returns a new Record with a
// Payload of NFC Forum external type.
func NewExternalRecord(extType string, payload []byte) *Record {
	pl := ext.New(extType, payload)
	return NewRecord(NFCForumExternalType, extType, "", pl)
}

// String a string representation of the payload of the record, prefixed
// by the URN of the resource.
//
// Note that not all NDEF Payloads are supported, and that custom types/payloads
// are considered not printable. In those cases, a generic RecordPayload is
// used and an explanatory message is returned instead.
// See submodules under "types/" for a list of supported types.
func (r *Record) String() string {
	pl, err := r.Payload()
	if err != nil {
		return err.Error()
	}
	return pl.Type() + ":" + pl.String()
}

// Inspect provides a string with information about this record.
// For a String representation of the contents use String().
func (r *Record) Inspect() string {
	if r.Empty() {
		return "Empty record"
	}

	pl, err := r.Payload()
	if err != nil {
		return err.Error()
	}

	var str string
	str += fmt.Sprintf("TNF: %d\n", r.TNF())
	str += fmt.Sprintf("Type: %s\n", r.Type())
	str += fmt.Sprintf("ID: %s\n", r.ID())
	str += fmt.Sprintf("MB: %t\n", r.MB())
	str += fmt.Sprintf("ME: %t\n", r.ME())
	str += fmt.Sprintf("Payload Length: %d", pl.Len())
	return str
}

// Unmarshal parses a byte slice into a Record struct (the slice can
// have extra bytes which are ignored). The Record is always reset before
// parsing.
//
// It does this by parsing every record chunk until a chunk with the CF flag
// cleared is read is read.
//
// Returns how many bytes were parsed from the slice (record length) or
// an error if something went wrong.
func (r *Record) Unmarshal(buf []byte) (rLen int, err error) {
	rLen = 0
	var chunks []*recordChunk
	for rLen < len(buf) {
		chunk := &recordChunk{}
		chunkSize, err := chunk.Unmarshal(buf[rLen:])
		rLen += chunkSize
		if err != nil {
			return rLen, err
		}
		chunks = append(chunks, chunk)

		// the last chunk record of a chunked record
		// has the CF flag cleared.
		if !chunk.CF {
			break
		}
	}

	r.chunks = chunks
	err = r.check()
	return rLen, err
}

// Marshal returns the byte representation of a Record. It does this
// by producing a single record chunk.
//
// Note that if the original Record was unmarshaled from many chunks,
// the recovery is not possible anymore.
func (r *Record) Marshal() ([]byte, error) {
	err := r.check()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for _, chunk := range r.chunks {
		chunkBytes, err := chunk.Marshal()
		if err != nil {
			return buf.Bytes(), err
		}
		_, err = buf.Write(chunkBytes)
		if err != nil {
			return buf.Bytes(), err
		}
	}
	return buf.Bytes(), nil
}

func (r *Record) check() error {
	chunksLen := len(r.chunks)
	last := chunksLen - 1
	if chunksLen == 0 {
		return errors.New(eNOCHUNKS)
	}
	if chunksLen == 1 && r.chunks[0].CF {
		return errors.New(eFIRSTCHUNKED)
	}
	if r.chunks[0].CF && r.chunks[last].CF {
		return errors.New(eLASTCHUNKED)
	}

	if chunksLen > 1 {
		chunksWithoutCF := 0
		chunksWithIL := 0
		chunksWithTypeLength := 0
		chunksWithoutUnchangedType := 0
		for i, r := range r.chunks {
			// Check CF in all but the last
			if !r.CF && i != last {
				chunksWithoutCF++
			}
			// Check IL in all but first
			if r.IL && i != 0 {
				chunksWithIL++
			}
			// TypeLength should be zero except in the first
			if r.TypeLength > 0 && i != 0 {
				chunksWithTypeLength++
			}
			// All but first chunk should have TNF to 0x06
			if r.TNF != Unchanged && i != 0 {
				chunksWithoutUnchangedType++
			}
		}
		if chunksWithoutCF > 0 {
			return errors.New(eCFMISSING)
		}
		if chunksWithIL > 0 {
			return errors.New(eBADIL)
		}
		if chunksWithTypeLength > 0 {
			return errors.New(eBADTYPELENGTH)
		}
		if chunksWithoutUnchangedType > 0 {
			return errors.New(eBADTNF)
		}
	}
	return nil
}

// Set some short-hands for the errors that can happen on check().
const (
	eNOCHUNKS      = "NDEF Record Check: No chunks"
	eFIRSTCHUNKED  = "NDEF Record Check: A single record cannot have the Chunk flag set"
	eLASTCHUNKED   = "NDEF Record Check: Last record cannot have the Chunk flag set"
	eCFMISSING     = "NDEF Record Check: Chunk Flag missing from some records"
	eBADIL         = "NDEF Record Check: IL flag is set on a middle or final chunk"
	eBADTYPELENGTH = "NDEF Record Check: A middle or last chunk has TypeLength != 0"
	eBADTNF        = "NDEF Record Check: A middle or last chunk TNF is not UNCHANGED"
)
