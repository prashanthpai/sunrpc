// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package sunrpc

import (
	"bytes"
	"encoding/binary"
	"io"
)

/*
From RFC 5531 (https://tools.ietf.org/html/rfc5531)

   A record is composed of one or more record fragments.  A record
   fragment is a four-byte header followed by 0 to (2**31) - 1 bytes of
   fragment data.  The bytes encode an unsigned binary number; as with
   XDR integers, the byte order is from highest to lowest.  The number
   encodes two values -- a boolean that indicates whether the fragment
   is the last fragment of the record (bit value 1 implies the fragment
   is the last fragment) and a 31-bit unsigned binary value that is the
   length in bytes of the fragment's data.  The boolean value is the
   highest-order bit of the header; the length is the 31 low-order bits.
*/

const (
	// This is maximum size in bytes for an individual record fragment.
	// The entire RPC message (record) has no size restriction imposed
	// by RFC 5531. Refer: include/linux/sunrpc/msg_prot.h
	maxRecordFragmentSize = (1 << 31) - 1
)

func isLastFragment(fragmentHeader uint32) bool {
	return (fragmentHeader >> 31) == 1
}

func getFragmentSize(fragmentHeader uint32) uint32 {
	return fragmentHeader &^ (1 << 31)
}

func createFragmentHeader(size uint32, lastFragment bool) uint32 {

	fragmentHeader := size &^ (1 << 31)

	if lastFragment {
		fragmentHeader |= (1 << 31)
	}

	return fragmentHeader
}

func minOf(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// WriteFullRecord writes the fully formed RPC message reply to network
// by breaking it into one or more record fragments.
func WriteFullRecord(conn io.Writer, data []byte) (int64, error) {

	dataSize := int64(len(data))
	dataReader := bytes.NewReader(data)

	var totalBytesWritten int64
	var lastFragment bool

	for {
		remainingBytes := dataSize - totalBytesWritten
		if remainingBytes <= maxRecordFragmentSize {
			lastFragment = true
		}
		fragmentSize := uint32(minOf(maxRecordFragmentSize, remainingBytes))

		// Create and write fragment header
		fragmentHeader := createFragmentHeader(fragmentSize, lastFragment)
		err := binary.Write(conn, binary.BigEndian, fragmentHeader)
		if err != nil {
			return totalBytesWritten, err
		}

		// Write fragment body (data) to network
		bytesWritten, err := io.CopyN(conn, dataReader, int64(fragmentSize))
		if err != nil || (bytesWritten != int64(fragmentSize)) {
			return totalBytesWritten, ErrWritingRecordFragment
		}
		totalBytesWritten += bytesWritten

		if lastFragment {
			break
		}
	}

	return totalBytesWritten, nil
}

// ReadFullRecord reads the entire RPC message from network and returns a
// a []byte sequence which contains the record.
func ReadFullRecord(conn io.Reader) ([]byte, error) {

	var fragmentHeader uint32
	record := new(bytes.Buffer)
	for {
		// Read record fragment header
		err := binary.Read(conn, binary.BigEndian, &fragmentHeader)
		if err != nil {
			return nil, err
		}

		fragmentSize := getFragmentSize(fragmentHeader)
		if fragmentSize > maxRecordFragmentSize {
			return nil, ErrInvalidFragmentSize
		}

		// Copy fragment body (data) from network to buffer
		bytesCopied, err := io.CopyN(record, conn, int64(fragmentSize))
		if err != nil || (bytesCopied != int64(fragmentSize)) {
			return nil, ErrReadingRecordFragment
		}

		if isLastFragment(fragmentHeader) {
			break
		}
	}

	return record.Bytes(), nil
}
