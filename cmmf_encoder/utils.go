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
	"fmt"
	"strconv"
	"strings"
)

// ceilDivision returns the ceil of `numerator`/`divisor`.
// It will return 0 if `divisor = 0`.
func ceilDivision(numerator, divisor int) int {
	if divisor == 0 {
		return 0
	}
	result := numerator / divisor
	if numerator%divisor != 0 {
		result++
	}
	return result
}

// writeUint32NumBytes will write `number` to the start of `buffer` using
// `numBytes` bytes in Big Endian.
//
// This function will panic if `buffer` is too small to accommodate `numBytes`.
//
// For example, `number = 1, numBytes = 1` will write []byte{0x01};
// `number = MaxUint32`, numBytes = 2` will write []byte{0xFF, 0xFF}.
func writeUint32NumBytes(buffer []byte, number uint32, numBytes int) {
	for idx := 0; idx < numBytes; idx++ {
		shift := (numBytes - (idx + 1)) * 8
		buffer[idx] = byte(number >> shift & 0xFF)
	}
}

// writeUnalignedBytes will join `existing` and `data` starting from bit
// `offset` of `existing`. The first `offset` bits of `existing` will be copied
// into the returned data; The remaining bits will be overwritten by bits in
// `data`.
//
// This function will always write complete bytes, meaning that if the offset
// you have in existing will be the offset you have in the last byte of the
// resulting data.
//
// `offset` is the bit offset within a byte, so it should be in the range [0, 7].
//
// Example: writeUnalignedBytes(0xab, 4, []byte{0x12, 0x34}) -> []byte{0xa1, 0x23, 0x40}.
//
// |-existing-|  |--------data-------|    |------------result-----------|
//
//	1010_1011    0001_0010 0011_0100  ->  1010_0001 0010_0011 0100_0000
//	     ^                               |----||------------------||---|
//	   offset                   (from existing) (from data)         (padding)
func writeUnalignedBytes(existing, offset byte, data []byte) []byte {
	if len(data) == 0 {
		return []byte{existing}
	}

	// Offset is the bit offset in a byte: convert values >=8.
	offset %= 8

	if offset == 0 {
		return data
	}

	result := make([]byte, len(data)+1)

	// We only want to keep the first `offset` bits from `existing`
	result[0] = existing & (0xff << (8 - offset))

	for idx := range data {
		left, right := splitByte(data[idx], offset)
		result[idx] |= left
		result[idx+1] = right
	}

	return result
}

// splitByte will partition `octet` into two parts.
// The `left` byte will contain the (8-`offset`) most significant bits; the
// `right` byte will contain the `offset` least significant bits.
//
// `offset` is the bit offset within a byte, so it should be in the range [0, 7].
//
// Example: splitByte(0xbd, 4) -> (0x0b, 0xd0)
func splitByte(octet, offset byte) (left, right byte) {
	// Offset is the bit offset in a byte: convert values >=8.
	offset %= 8

	offsetComplement := 8 - offset

	// The 1st byte should end with the first 'offset' bits of `octet`
	leftMask := byte(0xff << offset)
	left = (octet & leftMask) >> offset

	// The 2nd byte should start with the last 'offsetComplement' bits of `octet`
	rightMask := byte(0xff >> offsetComplement)
	right = (octet & rightMask) << offsetComplement

	return left, right
}

// parseBitString converts a configuration value written as a binary literal with a
// trailing "b", such as "011b", into its numeric value. The config schema uses this
// form for every fixed-width bitfield.
func parseBitString(s string) (uint8, error) {
	v, err := strconv.ParseUint(strings.TrimSuffix(s, "b"), 2, 8)
	if err != nil {
		return 0, fmt.Errorf("%q is not a binary literal of the form \"011b\": %w", s, err)
	}
	return uint8(v), nil
}
