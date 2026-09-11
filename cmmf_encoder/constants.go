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

const (
	// maxSubatomSize is the maximum number of bytes that any subatom (packet,
	// block header, bitstream header, etc.) can be. It includes the headers
	// specific to that subatom but does include the general subatom header.
	//
	// The spec allows for up to MaxUint32. Note that, to do this, we need a
	// 64-bit system.
	maxSubatomSize = (1 << 24) - 1

	// minSubatomHeaderSize is the smallest number of bytes that the non-data
	// portion of a subatom can be.
	minSubatomHeaderSize = 2

	// maxSubatomHeaderSize is the largest number of bytes that the non-data
	// portion of a subatom can be.
	maxSubatomHeaderSize = 5

	// minPcount is the minimum number of source symbols we can split a block
	// into.
	//
	// Since our primary purpose is multi-path, we set this to be >1.
	minPcount = 2

	// maxPcount is the maximum number of source symbols we can split a block
	// into.
	//
	// For now, we only use 1 byte to express num_symbols.
	// Hence, our range has a max of 255.
	maxPcount = 255

	// minBlockSize is the smallest number of bytes of original data that we
	// can support.
	//
	// This number is driven by the fact that we always try to create 2 source
	// symbols. We want each to carry some non-padding data.
	minBlockSize = 2

	// maxBlockSize is the largest number of bytes of original data that we
	// can support.
	//
	// The limit in the specification is MaxUint32 (due to the number of
	// bytes allowed to express 'block_size' in the block_header()).
	maxBlockSize = 1073732325

	// minPacketHeaderSize is the smallest possible packet_header in bytes.
	// This value does NOT include the subatom headers for a packet.
	minPacketHeaderSize = 2

	// maxPacketHeaderSize is the largest possible packet_header in bytes.
	// This value does NOT include the subatom headers for a packet.
	maxPacketHeaderSize = 33

	// minSymbolSize is the smallest number of bytes that can be in a source
	// symbol.
	//
	// This will also be the size of the smallest possible payload of a cmmf
	// packet.
	minSymbolSize = 1

	// maxSymbolSize is the largest number of bytes than can be in a source
	// symbol.
	//
	// The limit is from the fact that a packet must fit within a subatom, and
	// that a packet consists of a coded_symbol (the same size as a source
	// symbol) and some headers.
	// In practice, however, it is more likely to be limited by the blockSize.
	maxSymbolSize = maxSubatomSize - maxPacketHeaderSize

	// minPsize is the smallest possible 'psize' (that is, number of bytes in
	// a packet including subatom and packet headers).
	minPsize = minSubatomHeaderSize + minPacketHeaderSize + minSymbolSize

	// maxPsize is the largest possible 'psize' (that is, number of bytes in
	// a packet including subatom and packet headers).
	//
	// NOTE: '-1' because, since we adjusted the maxSubatomSize to be 2^24-1 to
	// be in line with the decoder, we will never need a full 4 bytes to express
	// the subatom size. Fow now, profile packets it will be 4.
	maxPsize = maxSubatomHeaderSize - 1 + maxPacketHeaderSize + maxSymbolSize

	// minPackedCoeff is the smallest number of bytes our (variable-length) packed
	// coefficient vector will take up.
	minPackedCoeff = 1
)

// CMMFSyncWord is the sync word we'll put at the beginning of bitstreams to
// signal that this is an ETSI CMMF bitstream.
// It translates as "\?x\rC\0D¬1" (where \? is UTF-9 char decimal 137).
//
// This is only a variable because Go will not allow it to be a constant.
var CMMFSyncWord = [8]byte{0x89, 0x78, 0x0d, 0x43, 0x00, 0x44, 0xac, 0x31}
