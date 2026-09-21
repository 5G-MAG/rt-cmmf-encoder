/******************************************************************************
 * 5G-MAG Reference Tools: CMMF Encoder
 ******************************************************************************
 * Copyright: (C)2024-2026 Dolby Laboratories Inc.
 * Author(s):
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

	"github.com/5G-MAG/rt-cmmf-encoder/cmmf_encoder/simd_xor"
)

type Packet struct {
	Data          []byte
	encoderConfig *ETSIConfig
}

func NewPacketFromEncoderConfig(psize int, encoderConfig *ETSIConfig) *Packet {
	return &Packet{
		Data:          make([]byte, psize),
		encoderConfig: encoderConfig,
	}
}

// writePacketHeader will write the `subatom` and `packet_header` headers into `pkt`.
// The coefficient_vector will be set to all zeros.
//
// The function will return the number of **bytes** written to the packet and
// the **bit** offset within the packet at which the coefficient vector starts.
// e.g. if the function returns (37, 39, nil), the 'coded_symbol' will start at
// pkt[37] and the coefficient_vector will begin at the 8th bit of the 4th byte.
//
// This function will error if the length of the packet is either too big or
// too small to support both sets of headers.
//
// This function currently only implements a subset of the ETSI CMMF profile.
func (pkt *Packet) writePacketHeader(pcount byte, symbolSize uint32) (int, int, error) {
	coeffVectorSizeBits := calculateCoeffVectorSize(pcount)
	packetHeaderSize := ceilDivision(8+coeffVectorSizeBits, 8)

	if tmpAtomSize := packetHeaderSize + int(symbolSize); tmpAtomSize > maxSubatomSize {
		return 0, 0, fmt.Errorf("packet is bigger than maximum allowed subatom (%d bytes)", tmpAtomSize)
	}
	packetAtomSize := uint32(packetHeaderSize) + symbolSize

	subatomHeaderLen := calculateSubatomHeaderSize(packetAtomSize)
	if lenPkt := len(pkt.Data); lenPkt < subatomHeaderLen+packetHeaderSize {
		return 0, 0, fmt.Errorf("packet is too small to fit full headers (%d bytes)", lenPkt)
	}

	offset := writeSubatomHeader(pkt.Data, SUBATOM_ID_PACKET, packetAtomSize)

	// For now, we will only set the coeff_vector part of the packet_mask.
	// All other settings in the packet_header will be 0.
	// This amounts to 8 bits: 1 bit to indicate whether systematic packet
	// (always set to 0) + 7 bits for the packet mask
	pkt.Data[offset] = 0x10

	// Next comes the coeff vector (leave as zeros), followed by a byte_align
	coeffVectorStartBit := offset*8 + 8
	bytesWritten := ceilDivision(coeffVectorStartBit+coeffVectorSizeBits, 8)

	return bytesWritten, coeffVectorStartBit, nil
}

// AddSourceSymbol encodes a source symbol (with an index of `symbolIdx` and
// containing `data`) into the packet's coded_symbol.
//
// `codedSymbolStartByte` is the **byte** offset at which the coded_symbol
// starts within `pkt`.
// `coeffVecStartBit` is the **bit** offset (from the start of the packet) at
// which the packet_header's coefficient_vector begins.
//
// If `data` has more bytes than `pkt`'s coded_symbol we will only add the first
// `len(pkt)` bytes. If it has equal or fewer bytes the whole symbol will be
// added.
//
// Each symbol can only be added once. If the given `symbolIdx` is already
// present we will error.
//
// This function currently only implements a subset of the ETSI CMMF profile.
func (pkt *Packet) AddSourceSymbol(codedSymbolStartByte, coeffVecStartBit int, symbolIdx byte, data []byte) error {
	if codedSymbolStartByte < 1 || codedSymbolStartByte >= len(pkt.Data) {
		return fmt.Errorf("no room for coded symbol: starts at byte %d of %d", codedSymbolStartByte, len(pkt.Data))
	} else if coeffVecEndByte := ceilDivision(coeffVecStartBit+int(symbolIdx+1), 8); coeffVecEndByte > codedSymbolStartByte {
		return fmt.Errorf(
			"coded_symbol starts at byte %d, but coefficient_vector ends at byte %d",
			codedSymbolStartByte, coeffVecEndByte,
		)
	}

	if err := pkt.addCoeffToVector(coeffVecStartBit, symbolIdx); err != nil {
		return fmt.Errorf("failed to add symbol %d to packet: %w", symbolIdx, err)
	}

	// encode the source symbol into the coded_symbol.
	simd_xor.XOR(pkt.Data[codedSymbolStartByte:], data)

	return nil
}

// addCoeffVector will the `symbolIdx`th symbol to the coefficient_vector.
//
// `coeffVecStartBit` is the bit offset from the start of `pkt` at which the
// coefficient_vector starts. We will error if the symbol is already present in
// the coefficient_vector, as well as for invalid args.
//
// This function currently only implements a subset of the ETSI CMMF profile.
// Specifically, the coefficient_vector is assumed to be GF2.
func (pkt *Packet) addCoeffToVector(coeffVecStartBit int, symbolIdx byte) error {
	if coeffVecStartBit < 0 {
		return fmt.Errorf("negative coefficient vector start bit: %d", coeffVecStartBit)
	}
	if symbolIdx < 0 || symbolIdx >= maxPcount {
		return fmt.Errorf("symbolIdx %d not in range [0, MaxPcount)", symbolIdx)
	}

	offset := coeffVecStartBit + int(symbolIdx)

	byteOffset := offset / 8
	bitOffsetWithinByte := offset % 8

	if byteOffset > len(pkt.Data)-1 {
		return fmt.Errorf("can't set bit in byte %d of %d", byteOffset, len(pkt.Data))
	}

	mask := byte(128) >> byte(bitOffsetWithinByte)

	if pkt.Data[byteOffset]&mask != 0x00 {
		return fmt.Errorf("symbol %d already in coeff vector", symbolIdx)
	}

	pkt.Data[byteOffset] |= mask
	return nil
}

// calculatePacketHeaderSize calculates the length of the non-coded_symbol
// portion of a cmmf packet for a given `pcount` and source `symbolSize`.
//
// This value will include the subatom header as well as the packet_header.
// We do not check that `pcount` and `symbolSize` have valid values. If they
// aren't valid, this function may return nonsense.
//
// This function currently only implements a subset of the ETSI CMMF profile.
func calculatePacketHeaderSize(pcount byte, symbolSize uint32) int {
	// We use 7 bits in packet_header portion plus the coefficient_vector.
	coeffVectorLen := calculateCoeffVectorSize(pcount)
	packetHeaderLen := ceilDivision(8+coeffVectorLen, 8)

	subatomHeaderLen := calculateSubatomHeaderSize(uint32(packetHeaderLen) + symbolSize)
	return subatomHeaderLen + packetHeaderLen
}

// calculateCoeffVectorSize returns the number of **bits** required to store the
// coefficient_vector portion of a cmmf packet_header(), given that the block
// has been split into `pcount` symbols.
//
// This function currently only implements a subset of the ETSI CMMF profile.
func calculateCoeffVectorSize(pcount byte) int {
	// For now, we assume GF2, each coefficient will take 1 bit.
	return int(pcount)
}

// calculatePsize returns the number of bytes required for each packet subatom
// (including subatom headers) given the coded symbol will by `symbolSize`
// bytes long and we've split the block into `pcount` symbols.
//
// This function currently only implements a subset of the ETSI CMMF profile.
func calculatePsize(pcount byte, symbolSize uint32) (psize int, err error) {
	psize = calculatePacketHeaderSize(pcount, symbolSize) + int(symbolSize)
	err = checkValidPsize(psize, pcount, symbolSize)
	return psize, err
}

// calculateSymbolSize returns the number of bytes in each source symbol given
// that we want to split a source block of `blockSize` bytes into `pcount`
// symbols.
//
// This function currently only implements a subset of the ETSI CMMF profile.
func calculateSymbolSize(pcount byte, blockSize uint32) (symbolSize uint32, err error) {
	if uint32(pcount) > blockSize {
		return 0, fmt.Errorf("pcount %d > blockSize %d", pcount, blockSize)
	}
	symbolSizeInt := ceilDivision(int(blockSize), int(pcount))
	err = checkValidSymbolSize(symbolSizeInt)
	symbolSize = uint32(symbolSizeInt)
	return symbolSize, err
}

func (pkt *Packet) Serialize() []byte {
	return pkt.Data
}
