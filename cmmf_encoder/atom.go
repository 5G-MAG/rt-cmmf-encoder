/******************************************************************************
 * 5G-MAG Reference Tools: CMMF Encoder
 ******************************************************************************
 * Copyright: (C)2024-2026 Dolby Laboratories Inc.
 * License: 5G-MAG Public License v1
 *
 * Licensed under the License terms and conditions for use, reproduction, and
 * distribution of 5G-MAG software (the “License”).  You may not use this file
 * except in compliance with the License.  You may obtain a copy of the License at
 * https://www.5g-mag.com/reference-tools.  Unless required by applicable law or
 * agreed to in writing, software distributed under the License is distributed on
 * an “AS IS” BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied.
 *
 * See the License for the specific language governing permissions and limitations
 * under the License.
 */

package cmmf_encoder

import (
	"math"
)

// writeSubatomHeader will write a subatom header to the start of `buffer`.
//
// `buffer` should be a byte slice into which the data will be written. The
// subatom header will take up the first 2-5 bytes.
// `subatomType` is the subatomID for this type of subatom.
// `subatomSize` should be the number of bytes the subatom will have (not
// including this header)
//
// The function will return the number of bytes written.
//
// This function will panic if `buffer` is too small to accommodate the header.
// The header will be between 2 and 5 bytes in length
//
// This function currently only implements a subset of the ETSI CMMF profile.
func writeSubatomHeader(buffer []byte, subatomType SubatomID, subatomSize uint32) int {
	sizeNumBytes := calculateBytesRequiredToStoreNum(subatomSize)

	// First 4 bits are the atom type
	buffer[0] = byte(subatomType) << 4

	// Last 2 bits of the first byte are the number of bytes the subatom_size
	// field will take.
	buffer[0] |= byte(sizeNumBytes - 1)

	writeUint32NumBytes(buffer[1:], subatomSize, sizeNumBytes)

	return 1 + sizeNumBytes
}

// getSubatomHeader will create and return a byte slice that contains a
// subatom header. The byte slice will contain only the header.
//
// `subatomType` is the subatomID for this type of subatom.
// `subatomSize` should be the number of bytes the subatom will have (not
// including this header)
//
// This function currently only implements a subset of the ETSI CMMF profile.
func getSubatomHeader(subatomType SubatomID, subatomSize uint32) (header []byte) {
	header = make([]byte, calculateSubatomHeaderSize(subatomSize))
	writeSubatomHeader(header, subatomType, subatomSize)
	return header
}

// calculateBytesRequiredToStoreNum returns the number of bytes required to store `number`.
func calculateBytesRequiredToStoreNum(number uint32) int {
	switch {
	case number <= math.MaxUint8:
		return 1
	case number <= math.MaxUint16:
		return 2
	case number <= 1<<24-1:
		return 3
	default:
		return 4
	}
}

// calculateSubatomHeaderSize returns the number of bytes required to write the
// subatom header.
//
// This function only supports a subset of the ETSI CMMF profile,
// so the only thing that will affect the header size will be the size of the
// subatom.
func calculateSubatomHeaderSize(subatomSize uint32) int {
	// The mandatory fields of the subatom take one byte.
	return 1 + calculateBytesRequiredToStoreNum(subatomSize)
}
