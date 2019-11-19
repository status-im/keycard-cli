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
	"encoding/binary"
	"errors"
	"fmt"
)

// BytesToUint64 parses a byte slice to an uint64 (BigEndian). If the slice
// is longer than 8 bytes, it's truncated. If it's shorter, they are considered
// the less significant bits of the uint64.
func bytesToUint64(b []byte) uint64 {
	// Make sure we are not parsing more than 8 bytes (uint64 size)
	byte8 := make([]byte, 8)
	if len(b) > 8 {
		copy(byte8, b[len(b)-8:]) // use the last 8 bytes
	} else {
		copy(byte8[8-len(b):], b) // copy to last positions of byte8
	}
	return binary.BigEndian.Uint64(byte8)
}

// Uint64ToBytes converts a BigEndian uint64 into a byte slice of
// desiredLen. For lengths under 8 bytes, the 8 byte result is
// truncated (the most significant bytes are discarded
func uint64ToBytes(n uint64, desiredLen int) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, n)
	if desiredLen >= 8 {
		slice := make([]byte, desiredLen)
		copy(slice[desiredLen-8:], buf.Bytes())
		return slice
	}

	return buf.Bytes()[8-desiredLen:]
}

// func PrintBytes(bytes []byte, length int) {
// 	for i := 0; i < length; i++ {
// 		fmt.Printf("%02x ", bytes[i])
// 	}
// 	fmt.Println()
// }

// FmtBytes receives a byte slice and a n value and returns
// a string with first n hex-formatted values of the slice
func fmtBytes(bytes []byte, n int) (str string) {
	if n > len(bytes) {
		n = len(bytes)
	}
	for i := 0; i < n; i++ {
		str += fmt.Sprintf("%02x ", bytes[i])
	}
	return str
}

// Reads from the bytes buffer and panics with an error
// if the buffer does not have the bytes to read that we want
// This is meant as a replacement for byteslice[3:3+x] where
// I don't have to constantly check the array length
// but instead I throw a custom panic
func getBytes(b *bytes.Buffer, n int) []byte {
	slice := make([]byte, n)
	nread, err := b.Read(slice)
	if err != nil || nread != n {
		panic(errors.New("unexpected end of data"))
	}
	return slice
}

// Same as above bug for a single byte
func getByte(b *bytes.Buffer) byte {
	byte, err := b.ReadByte()
	if err != nil {
		panic(errors.New("unexpected end of data"))
	}
	return byte
}
